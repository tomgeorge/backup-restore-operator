package volumebackup

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"testing"

	helpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func int32Pointer(n int32) *int32 {
	return &n
}

// Test null volumes
// What happens when you backup an application that has no volume
func TestReconcile(t *testing.T) {
	var (
		name                 = "example-backup"
		namespace            = "example"
		applicationRef       = "example-application-to-backup"
		replicas       int32 = 1
		gvr                  = schema.GroupVersionResource{
			Group:    "volumesnapshot",
			Resource: "volumesnapshots",
			Version:  "v1alpha1",
		}
	)

	cases := []testCase{
		{
			name: "reconcile add - one container with one volume",
			objs: []runtime.Object{
				helpers.NewDeployment(namespace, applicationRef, &replicas),
				helpers.NewPod(namespace, applicationRef, 1),
				helpers.NewPersistentVolume("test-vol-0"),
				helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: helpers.NewVolumeBackup(namespace, name, applicationRef),
			expectedActions: []core.Action{
				core.NewCreateAction(gvr,
					namespace,
					helpers.NewVolumeSnapshot(namespace, applicationRef+"-data-0", "test-claim-0"),
				),
			},
		},
		{
			name: "reconcile add - one container that has multiple volumes",
			objs: []runtime.Object{
				helpers.NewDeployment(namespace, applicationRef, &replicas),
				helpers.NewPod(namespace, applicationRef, 2),
				helpers.NewPersistentVolume("test-vol-0"),
				helpers.NewPersistentVolume("test-vol-1"),
				helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
				helpers.NewPersistentVolumeClaim(namespace, "test-claim-1", "test-vol-1"),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: helpers.NewVolumeBackup(namespace, name, applicationRef),
			expectedActions: []core.Action{
				core.NewCreateAction(
					gvr,
					namespace,
					helpers.NewVolumeSnapshot(namespace, applicationRef+"-data-0", "test-claim-0"),
				),
				core.NewCreateAction(
					gvr,
					namespace,
					helpers.NewVolumeSnapshot(namespace, applicationRef+"-data-1", "test-claim-1"),
				),
			},
		},
		{
			name: "reconcile add - VolumeBackup object has not been created yet",
			objs: []runtime.Object{
				helpers.NewDeployment(namespace, applicationRef, &replicas),
				helpers.NewPod(namespace, applicationRef, 1),
				helpers.NewPersistentVolume("test-vol-0"),
				helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs:    []runtime.Object{},
			expectedActions: []core.Action{},
			expectedResult: reconcile.Result{
				Requeue:      false,
				RequeueAfter: 0,
			},
			expectedError: nil,
		},
		{
			name:            "reconcile add - can't find deployment",
			objs:            []runtime.Object{},
			snapshotObjs:    []runtime.Object{},
			volumeBackup:    helpers.NewVolumeBackup(namespace, name, applicationRef),
			expectedActions: []core.Action{},
			expectedResult:  reconcile.Result{},
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "apps",
				Resource: "deployments",
			}, applicationRef),
		},
		{
			name: "reconcile add - no pods available - basically a no-op",
			objs: []runtime.Object{
				helpers.NewDeployment(namespace, applicationRef, &replicas),
				helpers.NewPersistentVolume("test-vol-0"),
				helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs:    []runtime.Object{},
			volumeBackup:    helpers.NewVolumeBackup(namespace, name, applicationRef),
			expectedActions: []core.Action{},
			expectedResult:  reconcile.Result{},
		},
	}

	for _, testCase := range cases {
		if !testCase.skip {
			runInTestHarness(t, testCase)
		}
	}
}
