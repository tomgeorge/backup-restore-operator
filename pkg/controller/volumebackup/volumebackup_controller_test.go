package volumebackup

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"testing"

	test_helpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
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
		name                  = "example-backup"
		namespace             = "example"
		applicationName       = "example-application-to-backup"
		containerName         = "busybox"
		replicas        int32 = 1
		gvr                   = schema.GroupVersionResource{
			Group:    "volumesnapshot",
			Resource: "volumesnapshots",
			Version:  "v1alpha1",
		}
	)

	cases := []testCase{
		{
			name: "reconcile add - one container with one volume",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPod(namespace, applicationName, 1),
				test_helpers.NewPersistentVolume("test-vol-0"),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{
				core.NewCreateAction(gvr,
					namespace,
					test_helpers.NewVolumeSnapshot(namespace, applicationName+"-data-0", "test-claim-0"),
				),
			},
		},
		{
			name: "reconcile add - one container that has multiple volumes",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPod(namespace, applicationName, 2),
				test_helpers.NewPersistentVolume("test-vol-0"),
				test_helpers.NewPersistentVolume("test-vol-1"),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-1", "test-vol-1"),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{
				core.NewCreateAction(
					gvr,
					namespace,
					test_helpers.NewVolumeSnapshot(namespace, applicationName+"-data-0", "test-claim-0"),
				),
				core.NewCreateAction(
					gvr,
					namespace,
					test_helpers.NewVolumeSnapshot(namespace, applicationName+"-data-1", "test-claim-1"),
				),
			},
		},
		{
			name: "reconcile add - VolumeBackup object has not been created yet",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPod(namespace, applicationName, 1),
				test_helpers.NewPersistentVolume("test-vol-0"),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
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
			volumeBackup:    test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{},
			expectedResult:  reconcile.Result{},
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "apps",
				Resource: "deployments",
			}, applicationName),
		},
		{
			name: "reconcile add - no pods available - basically a no-op",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPersistentVolume("test-vol-0"),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs:    []runtime.Object{},
			volumeBackup:    test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{},
			expectedResult:  reconcile.Result{Requeue: true},
		},
		{
			name: "reconcile add - volume does not exist",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPod(namespace, applicationName, 1),
				test_helpers.NewPersistentVolumeClaim(namespace, "test-claim-0", "test-vol-0"),
			},
			snapshotObjs:    []runtime.Object{},
			volumeBackup:    test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{},
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "persistentvolumes",
			}, "test-vol-0"),
		},
		{
			name: "reconcile add - volume claim does not exist",
			objs: []runtime.Object{
				test_helpers.NewDeployment(namespace, applicationName, &replicas),
				test_helpers.NewPod(namespace, applicationName, 1),
				test_helpers.NewPersistentVolume("test-vol-0"),
			},
			snapshotObjs:    []runtime.Object{},
			volumeBackup:    test_helpers.NewVolumeBackup(namespace, name, applicationName, containerName, "test-vol-0"),
			expectedActions: []core.Action{},
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "persistentvolumeclaims",
			}, "test-claim-0"),
		},
	}

	for _, testCase := range cases {
		if !testCase.skip {
			runInTestHarness(t, testCase)
		}
	}
}
