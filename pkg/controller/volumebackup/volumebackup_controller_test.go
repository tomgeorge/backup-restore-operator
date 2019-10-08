package volumebackup

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	testHelpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
)

func int32Pointer(n int32) *int32 {
	return &n
}

func TestReconcile(t *testing.T) {
	var (
		name                  = "example-backup"
		namespace             = "example"
		applicationName       = "example-application-to-backup"
		volumeName            = "test-vol-0"
		claimName             = "test-claim-0"
		containerName         = "busybox"
		replicas        int32 = 1
		gvr                   = schema.GroupVersionResource{
			Group:    "snapshot.storage.k8s.io",
			Resource: "volumesnapshots",
			Version:  "v1alpha1",
		}
		statusEmpty     = &backupsv1alpha1.VolumeBackupStatus{}
		statusPodFrozen = &backupsv1alpha1.VolumeBackupStatus{
			PodPhase: backupsv1alpha1.PhaseFrozen,
			Conditions: []backupsv1alpha1.VolumeBackupCondition{
				{
					Type:   backupsv1alpha1.PodFrozen,
					Status: backupsv1alpha1.ConditionTrue,
				},
			},
		}
		statusSnapshotIssued = &backupsv1alpha1.VolumeBackupStatus{
			PodPhase: backupsv1alpha1.PhaseFrozen,
			Conditions: []backupsv1alpha1.VolumeBackupCondition{
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
			PodPhase: backupsv1alpha1.PhaseFrozen,
			Conditions: []backupsv1alpha1.VolumeBackupCondition{
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
		statusUploading = &backupsv1alpha1.VolumeBackupStatus{
			PodPhase: backupsv1alpha1.PhaseFrozen,
			Conditions: []backupsv1alpha1.VolumeBackupCondition{
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
				{
					Type:   backupsv1alpha1.SnapshotUploading,
					Status: backupsv1alpha1.ConditionTrue,
				},
			},
		}
	)

	cases := []testCase{
		{
			name: "no phase - should identify that the backup flow has not started, freeze the pod, and move the backup into the PodFrozen phase",
			objs: []runtime.Object{
				testHelpers.NewDeployment(namespace, applicationName, &replicas),
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs:         []runtime.Object{},
			volumeBackup:         testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusEmpty),
			expectedResult:       reconcile.Result{},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusPodFrozen),
			expectedKubeActions: []core.Action{
				core.NewUpdateAction(gvr, namespace, testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName, statusEmpty)),
			},
			expectedSnapshotActions: []core.Action{},
		},
		{
			name: "no phase - should not freeze if things are missing",
			objs: []runtime.Object{
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs:         []runtime.Object{},
			volumeBackup:         testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusEmpty),
			expectedKubeActions:  []core.Action{},
			expectedResult:       reconcile.Result{},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusEmpty),
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "apps",
				Resource: "deployments",
			}, applicationName),
		},
		{
			name: "Pod is frozen, should create a VolumeSnapshot object, and update the VolumeBackup status to SnapshotIssued",
			objs: []runtime.Object{
				testHelpers.NewDeployment(namespace, applicationName, &replicas),
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusPodFrozen),
			expectedSnapshotActions: []core.Action{
				core.NewCreateAction(gvr, namespace, testHelpers.NewVolumeSnapshot(namespace, "example-application-to-backup-test-claim-0", claimName)),
			},
			expectedVolumeBackup:   testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotIssued),
			expectedVolumeSnapshot: testHelpers.NewVolumeSnapshot(namespace, "example-application-to-backup-test-claim-0", claimName),
			expectedResult: reconcile.Result{
				Requeue: false,
			},
			expectedError: nil,
		},
		{
			name: "SnapshotIssued, should wait for a ReadyToUse/CreationTimestamp on the Snapshot and then update the status to SnapshotCreated",
			objs: []runtime.Object{
				testHelpers.NewDeployment(namespace, applicationName, &replicas),
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs: []runtime.Object{
				testHelpers.NewReadyVolumeSnapshot(namespace, "example-application-to-backup-test-claim-0", claimName),
			},
			volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotIssued),
			expectedKubeActions: []core.Action{
				core.NewUpdateAction(gvr, namespace, testHelpers.NewVolumeBackup(namespace, "example-application-to-backup-test-claim-0", applicationName, containerName, claimName, statusSnapshotIssued)),
			},
			expectedSnapshotActions: []core.Action{
				core.NewGetAction(gvr, namespace, "example-application-to-backup-test-claim-0"),
			},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotCreated),
			expectedResult: reconcile.Result{
				Requeue: false,
			},
			expectedError: nil,
		},
		{
			name: "SnapshotIssued, should wait for a ReadyToUse/CreationTimestamp on the Snapshot and then update the status to SnapshotCreated",
			objs: []runtime.Object{
				testHelpers.NewDeployment(namespace, applicationName, &replicas),
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs: []runtime.Object{
				testHelpers.NewReadyVolumeSnapshot(namespace, "example-application-to-backup-test-claim-0", claimName),
			},
			volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotIssued),
			expectedKubeActions: []core.Action{
				core.NewUpdateAction(gvr, namespace, testHelpers.NewVolumeBackup(namespace, "example-application-to-backup-test-claim-0", applicationName, containerName, claimName, statusSnapshotIssued)),
			},
			expectedSnapshotActions: []core.Action{
				core.NewGetAction(gvr, namespace, "example-application-to-backup-test-claim-0"),
			},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotCreated),
			expectedResult: reconcile.Result{
				Requeue: false,
			},
			expectedError: nil,
		},
		{
			name: "SnapshotCreated, should create a pod and like upload it and shit",
			objs: []runtime.Object{
				testHelpers.NewDeployment(namespace, applicationName, &replicas),
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs: []runtime.Object{
				testHelpers.NewReadyVolumeSnapshot(namespace, "example-application-to-backup-test-claim-0", claimName),
			},
			volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusSnapshotCreated),
			expectedKubeActions: []core.Action{
				core.NewCreateAction(gvr, namespace, &corev1.Pod{}),
			},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, claimName, statusUploading),
			expectedResult: reconcile.Result{
				Requeue: false,
			},
			expectedError: nil,
		},
	}

	for _, testCase := range cases {
		if !testCase.skip {
			runInTestHarness(t, testCase)
		}
	}
}
