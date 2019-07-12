package volumebackup

import (
	"context"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	fakeSnapshotClient "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/fake"
	snapshotscheme "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/scheme"
	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	core "k8s.io/client-go/testing"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type testCase struct {
	// The test case name, this will show up in the test logs
	name string

	// The runetime objects that the test expects to be in the cluster already.
	// For example, to test a create action of a volumebackup, you would need a deployment with some pods.
	// The deployment and pod(s) would go in objs.
	objs []runtime.Object

	// The VolumeSnapshot objects that the test expects to be in the cluster already.
	// If you were testing a delete action, you would add a VolumeSnapshot object, and then
	// issue a delete request in the test
	snapshotObjs []runtime.Object

	// A VolumeBackup object that is expected to be there when the test runs
	volumeBackup *volumebackupv1alpha1.VolumeBackup

	// The expected actions (Create/Update/Delete) that the snapshot client is supposed to perform
	expectedActions []core.Action

	// Should we skip this test?  Just for debugging purposes, please dont skip tests
	skip bool
}

func runInTestHarness(t *testing.T, test testCase) {
	snapshotscheme.AddToScheme(scheme.Scheme)
	volumebackupv1alpha1.AddToScheme(scheme.Scheme)
	t.Logf("Running test case %s", test.name)

	k8sClient := fakeClient.NewFakeClientWithScheme(scheme.Scheme, test.objs...)
	snapClientset := fakeSnapshotClient.NewSimpleClientset(test.snapshotObjs...)
	cfg := &rest.Config{}

	err := k8sClient.Create(context.TODO(), test.volumeBackup)
	if err != nil {
		t.Errorf("Error creating VolumeBackup: %v", err)
	}

	reconcileVolumeBackup := &ReconcileVolumeBackup{
		scheme:        scheme.Scheme,
		client:        k8sClient,
		config:        cfg,
		snapClientset: snapClientset,
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: test.volumeBackup.GetNamespace(),
			Name:      test.volumeBackup.GetName(),
		},
	}

	result, err := reconcileVolumeBackup.Reconcile(request)
	if err != nil {
		t.Errorf("Error reconciling object: %v", err)
	}

	if !result.Requeue {
		t.Logf("Reconcile did not requeue request as expected")
	}

	evaluateResults(test, reconcileVolumeBackup, t)
}

func evaluateResults(testcase testCase, reconcileVolumeBackup *ReconcileVolumeBackup, t *testing.T) {
	client, ok := reconcileVolumeBackup.snapClientset.(*fakeSnapshotClient.Clientset)
	if !ok {
		t.Errorf("Fatal - test %v - could not assert fakeSnapshotClient.Clientset type on snapshot client", testcase.name)
	}

	if len(client.Actions()) != len(testcase.expectedActions) {
		t.Errorf("Error - test %v - expected %v actions received by client but was %v.  The test case is probably misconfigured", testcase.name, len(testcase.expectedActions), len(client.Actions()))
		t.FailNow()
	}
	for index, expected := range testcase.expectedActions {
		actual := client.Actions()[index]
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Error - test %v - objects do not match", testcase.name)
			t.Errorf("%v", pretty.Diff(expected, actual))
		}
	}
}
