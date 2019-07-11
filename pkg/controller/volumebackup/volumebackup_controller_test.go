package volumebackup

import (
	"testing"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
)

func int32Pointer(n int32) *int32 {
	return &n
}

// Things to test:
// multiple backup requests
// what happens when the freeze/unfreeze command is bad?
// deletion of a backup, does it delete every snapshot?
//
// monotonically increasing backup names
// if you go this route, should the heirarchy be
// - ApplicationBackup, says what you want to back up
// - BackupRequest, request for backup
// - each backup request corresponds to e.g. `app-backup-mysql-data-1`
func TestReconcile(t *testing.T) {
	var (
		name                 = "example-backup"
		namespace            = "example"
		applicationRef       = "example-application-to-backup"
		storageclass         = "mock-csi-storageclass"
		replicas       int32 = 1
	)

	cases := []testCase{
		{
			name: "reconcile add",
			objs: []runtime.Object{
				newDeployment(namespace, applicationRef, &replicas),
				newPod(namespace, applicationRef),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: newVolumeBackup(namespace, name, applicationRef, storageclass),
			expectedActions: []core.Action{
				core.CreateActionImpl{
					ActionImpl: core.ActionImpl{
						Namespace: namespace,
						Verb:      "create",
						Resource: schema.GroupVersionResource{
							Group:    "volumesnapshot",
							Resource: "volumesnapshots",
							Version:  "v1alpha1",
						},
					},
					Name: "",
					Object: &v1alpha1.VolumeSnapshot{
						ObjectMeta: v1.ObjectMeta{
							Name:      "snapshot-example-application-to-backup-pod-volume-test-claim",
							Namespace: "example",
						},
						Spec: v1alpha1.VolumeSnapshotSpec{
							VolumeSnapshotClassName: &storageclass,
							Source: &corev1.TypedLocalObjectReference{
								Kind: "PersistentVolumeClaim",
								Name: "test-claim",
							},
						},
					},
				},
			},
		},
	}
	testSuite(t, cases)
}
