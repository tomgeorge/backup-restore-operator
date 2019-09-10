package volumebackup

import (
	"testing"

	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	name                  = "example-backup"
	namespace             = "example"
	applicationName       = "example-application-to-backup"
	volumeName            = "test-vol-0"
	claimName             = "test-claim-0"
	containerName         = "busybox"
	replicas        int32 = 1
	gvr                   = schema.GroupVersionResource{
		Group:    "volumesnapshot",
		Resource: "volumesnapshots",
		Version:  "v1alpha1",
	}
	statusEmpty     = &backupsv1alpha1.VolumeBackupStatus{}
	statusPodFrozen = &backupsv1alpha1.VolumeBackupStatus{
		VolumeBackupConditions: []backupsv1alpha1.VolumeBackupCondition{
			{
				Type:   backupsv1alpha1.PodFrozen,
				Status: backupsv1alpha1.ConditionTrue,
			},
		},
	}
	statusSnapshotIssued = &backupsv1alpha1.VolumeBackupStatus{
		VolumeBackupConditions: []backupsv1alpha1.VolumeBackupCondition{
			{
				Type:   backupsv1alpha1.PodFrozen,
				Status: backupsv1alpha1.ConditionTrue,
			},
			{
				Type:   backupsv1alpha1.SnapshotIssued,
				Status: backupsv1alpha1.ConditionTrue,
			},
		},
	}
	statusSnapshotCreated = &backupsv1alpha1.VolumeBackupStatus{
		VolumeBackupConditions: []backupsv1alpha1.VolumeBackupCondition{
			{
				Type:   backupsv1alpha1.PodFrozen,
				Status: backupsv1alpha1.ConditionTrue,
			},
			{
				Type:   backupsv1alpha1.SnapshotIssued,
				Status: backupsv1alpha1.ConditionTrue,
			},
			{
				Type:   backupsv1alpha1.SnapshotCreated,
				Status: backupsv1alpha1.ConditionTrue,
			},
		},
	}
)

cases := []newTestCase {

}

func TestReconcile(t *testing.T) {
	for _, testCase := range cases {
		if !testCase.skip {
			runInNewTestHarness(t, testCase)
		}
	}
}
