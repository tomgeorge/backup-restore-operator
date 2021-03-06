package volumebackup

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	fakeSnapshotClient "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/fake"
	snapshotscheme "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned/scheme"
	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"github.com/tomgeorge/backup-restore-operator/pkg/util/executor"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	core "k8s.io/client-go/testing"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// testCase contains options for the run of a test
type testCase struct {

	// The test case name, this will show up in the test logs
	name string

	// The runtime objects that the test expects to be in the cluster already.
	// For example, to test a create action of a volumebackup, you would need a deployment with some pods.
	// The deployment and pod(s) would go in objs.
	objs []runtime.Object

	// The VolumeSnapshot objects that the test expects to be in the cluster already.
	// If you were testing a delete action, you would add a VolumeSnapshot object, and then
	// issue a delete request in the test
	snapshotObjs []runtime.Object

	// A VolumeBackup object that is expected to be there when the test runs
	volumeBackup *volumebackupv1alpha1.VolumeBackup

	// The expected actions (Create/Update/Delete) that the kubernetes client is supposed to perform
	expectedKubeActions []core.Action

	// The expected actions to be performed by the snapshot client set
	expectedSnapshotActions []core.Action

	// The VolumeBackup expected after a reconcile in a given state
	expectedVolumeBackup *backupsv1alpha1.VolumeBackup

	// The expected VolumeSnapshot after the test is run
	expectedVolumeSnapshot *v1alpha1.VolumeSnapshot

	// The expected reconcile result
	expectedResult reconcile.Result

	// The expected errors that should occur during a test case
	expectedError error

	// Should this test cause a requeue?
	requeue bool

	// Should we skip this test?  Just for debugging purposes, please dont skip tests
	skip bool
}

func runInTestHarness(t *testing.T, test testCase) {
	logf.SetLogger(zap.LoggerTo(os.Stdout, true))
	snapshotscheme.AddToScheme(scheme.Scheme)
	volumebackupv1alpha1.AddToScheme(scheme.Scheme)
	backupsv1alpha1.AddToScheme(scheme.Scheme)
	t.Logf("Running test case %s", test.name)

	k8sClient := fakeClient.NewFakeClientWithScheme(scheme.Scheme, test.objs...)
	snapClientset := fakeSnapshotClient.NewSimpleClientset(test.snapshotObjs...)

	cfg := &rest.Config{}
	executor := executor.CreateNewFakePodExecutor()
	reconcileVolumeBackup := &ReconcileVolumeBackup{
		scheme:        scheme.Scheme,
		client:        k8sClient,
		config:        cfg,
		snapClientset: snapClientset,
		executor:      executor,
	}

	request := reconcile.Request{}

	// In some cases, you will want to test Reconciliation without a VolumeBackup having been
	// created
	if test.volumeBackup != nil {
		err := k8sClient.Create(context.TODO(), test.volumeBackup)
		if err != nil {
			t.Errorf("Error creating VolumeBackup: %v", err)
		}

		request.NamespacedName = types.NamespacedName{
			Namespace: test.volumeBackup.Namespace,
			Name:      test.volumeBackup.Name,
		}

	} else {
		fmt.Printf("This test does not have a volumebackup in the set-up")
		request.NamespacedName = types.NamespacedName{
			Namespace: "default",
			Name:      "vb",
		}
	}
	result, err := reconcileVolumeBackup.Reconcile(request)
	evaluateResults(test, reconcileVolumeBackup, result, err, t)
}

func evaluateResults(testcase testCase, reconcileVolumeBackup *ReconcileVolumeBackup, result reconcile.Result, err error, t *testing.T) {
	snapshotClient, ok := reconcileVolumeBackup.snapClientset.(*fakeSnapshotClient.Clientset)
	if !ok {
		t.Errorf("Fatal - test %v - could not assert fakeSnapshotClient.Clientset type on snapshot client", testcase.name)
	}

	// Fail if the expected actions are not the same as the actual actions in the snapshot clientset
	if len(snapshotClient.Actions()) != len(testcase.expectedSnapshotActions) {
		t.Errorf("Error - test %v - expected %v actions received by client but was %v", testcase.name, len(testcase.expectedSnapshotActions), len(snapshotClient.Actions()))
		t.Errorf("%v", pretty.Diff(testcase.expectedSnapshotActions, snapshotClient.Actions()))
		t.FailNow()
	}
	for index, expected := range testcase.expectedSnapshotActions {
		actual := snapshotClient.Actions()[index]
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Error - test %v - objects do not match", testcase.name)
			t.Errorf("%v", pretty.Diff(expected, actual))
		}
	}

	// Fail if the expected error does not match the actual error
	if !reflect.DeepEqual(err, testcase.expectedError) {
		t.Errorf("Error - test %v - expected error but got", testcase.name)
		t.Errorf("%v", pretty.Diff(err, testcase.expectedError))
	}

	// Fail if the expected reconcile.Result is not the same as the actual reconcile.Result
	if !reflect.DeepEqual(result, testcase.expectedResult) {
		t.Errorf("Error - test %v - result objects do not match", testcase.name)
		t.Errorf("%v", pretty.Diff(result, testcase.expectedResult))
	}

	// Request the VolumeBackup object from the client, and fail if the status is not the same as the expected volumebackup
	if testcase.expectedVolumeBackup != nil {
		backup := &backupsv1alpha1.VolumeBackup{}
		getErr := reconcileVolumeBackup.client.Get(context.TODO(), types.NamespacedName{
			Namespace: testcase.expectedVolumeBackup.Namespace,
			Name:      testcase.expectedVolumeBackup.Name,
		}, backup)
		if getErr != nil {
			t.Errorf("Error - test %v - could not get volumebackup", testcase.name)
		}

		for i, condition := range backup.Status.Conditions {
			if !(testcase.expectedVolumeBackup.Status.Conditions[i].Type == condition.Type &&
				testcase.expectedVolumeBackup.Status.Conditions[i].Status == condition.Status) {
				t.Errorf("Error - test %v - expected backup status does not match actual", testcase.name)
				t.Errorf("%v", pretty.Diff(testcase.expectedVolumeBackup.Status.Conditions, backup.Status.Conditions))
			}
		}
	}

	if testcase.expectedVolumeSnapshot != nil {
		actualVolumeSnapshot, err := snapshotClient.SnapshotV1alpha1().
			VolumeSnapshots(testcase.expectedVolumeSnapshot.Namespace).
			Get(testcase.expectedVolumeSnapshot.Name, v1.GetOptions{})

		if err != nil {
			t.Errorf("Error - test %v - can't get VolumeSnapshot: %v", testcase.name, err)
		}

		if !reflect.DeepEqual(actualVolumeSnapshot, testcase.expectedVolumeSnapshot) {
			t.Errorf("Error - test %v - actual object and expected are not the same", testcase.name)
			t.Errorf("%v", pretty.Diff(actualVolumeSnapshot, testcase.expectedVolumeSnapshot))
		}

	}

	// Compare the expected volumesnapshot (if any) against the one returned by the API server
}
