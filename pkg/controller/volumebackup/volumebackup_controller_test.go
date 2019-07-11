package volumebackup

import (
	"testing"

	helpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
)

func int32Pointer(n int32) *int32 {
	return &n
}

func TestReconcile(t *testing.T) {
	var (
		name                 = "example-backup"
		namespace            = "example"
		applicationRef       = "example-application-to-backup"
		storageclass         = "mock-csi-storageclass"
		replicas       int32 = 1
		gvr                  = schema.GroupVersionResource{
			Group:    "volumesnapshot",
			Resource: "volumesnapshots",
			Version:  "v1alpha1",
		}
	)

	cases := []testCase{
		{
			name: "reconcile add",
			objs: []runtime.Object{
				helpers.NewDeployment(namespace, applicationRef, &replicas),
				helpers.NewPod(namespace, applicationRef, 1),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: helpers.NewVolumeBackup(namespace, name, applicationRef, storageclass),
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
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: helpers.NewVolumeBackup(namespace, name, applicationRef, storageclass),
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
	}

	for _, testCase := range cases {
		runInTestHarness(t, testCase)
	}
}
