package volumebackup

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	snapshotscheme "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/scheme"
	"k8s.io/client-go/kubernetes/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	core "k8s.io/client-go/testing"

	fakeSnapshotClient "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/fake"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"

	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type fixture struct {
	t *testing.T

	client      *fake.Clientset
	kubeclient  *k8sfake.Clientset
	kubeactions []core.Action
	actions     []core.Action
	kubeObjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeObjects = []runtime.Object{}
	return f
}

// Sanity test that checks that you can create a VolumeBackup and get back
// a VolumeBackupList of one element
func TestCanCreateABackupForApplication(t *testing.T) {
	volumebackup := newVolumeBackup(metav1.NamespaceDefault, "my-backup", "my-application")
	objs := []runtime.Object{volumebackup, newVolumeBackupList()}
	s := scheme.Scheme
	s.AddKnownTypes(volumebackupv1alpha1.SchemeGroupVersion, objs...)
	client := fakeClient.NewFakeClientWithScheme(s, objs...)
	backups := &volumebackupv1alpha1.VolumeBackupList{}
	err := client.List(context.TODO(), &runtimeClient.ListOptions{Namespace: "default"}, backups)
	if err != nil {
		t.Errorf("Error listing VolumeBackups: %v", err)
	}

	for e, index := range backups.Items {
		t.Logf("backup %v=%v", index, e)
	}
	if len(backups.Items) != 1 {
		fmt.Printf("%v", backups.Items)
		t.Errorf("Expected 1 element but had %v", len(backups.Items))
	}
}

// Tests that when a VolumeBackup is created, the reconciliation loop will also create
func TestReconcileAdd(t *testing.T) {
	var (
		name           = "example-backup"
		namespace      = "example"
		applicationRef = "example-application-to-backup"
		replicas       int32
	)

	snapshotscheme.AddToScheme(scheme.Scheme)
	volumebackupv1alpha1.AddToScheme(scheme.Scheme)

	volumebackup := newVolumeBackup(namespace, name, applicationRef)
	application, pod := newDeploymentAndPod(namespace, applicationRef, &replicas)

	objs := []runtime.Object{}
	k8sClient := fakeClient.NewFakeClientWithScheme(scheme.Scheme, objs...)

	snapObjs := []runtime.Object{}
	snapClient := fakeSnapshotClient.NewSimpleClientset(snapObjs...)

	cfg := &rest.Config{}

	err := k8sClient.Create(context.TODO(), application)
	if err != nil {
		t.Errorf("Error createing deployment: %v", err)
	}
	t.Logf("After k8sClientCreate deployment")

	err = k8sClient.Create(context.TODO(), pod)
	if err != nil {
		t.Errorf("Error createing pod: %v", err)
	}
	t.Logf("After k8sClientCreate pod")

	err = k8sClient.Create(context.TODO(), volumebackup)

	r := &ReconcileVolumeBackup{
		scheme:        scheme.Scheme,
		snapClientset: snapClient,
		client:        k8sClient,
		config:        cfg,
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}

	result, err := r.Reconcile(request)

	t.Logf("After reconcile call")

	if !result.Requeue {
		t.Logf("Reconcile did not requeue request as expected")
	}

	if len(snapClient.Actions()) != 1 {
		t.Errorf("Expecting snapClient.Actions to have length 1, was %v", len(snapClient.Actions()))
		t.FailNow()
	}

	action := snapClient.Actions()[0]
	t.Logf("action is %v", action)

	groupVersionResource := v1alpha1.Resource("volumesnapshots").WithVersion("v1alpha1")
	t.Logf("gvr %v", groupVersionResource)
	if action.GetNamespace() != namespace ||
		action.GetVerb() != "create" ||
		action.GetResource().Resource != "volumesnapshots" {
		t.Errorf("Expected create snapshot action in namespace %v. Instead got namespace %v verb %v resource %v", namespace, action.GetNamespace(), action.GetVerb(), action.GetResource())

	}

	if err != nil {
		t.Errorf("Error fetching volumesnapshots in namespace %v: %v", namespace, err)
	}
}
