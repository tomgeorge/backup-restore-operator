package volumebackup

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	testHelpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
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
			snapshotObjs: []runtime.Object{},
			volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName, statusEmpty),
			expectedActions: []core.Action{
				core.NewGetAction(gvr,
					namespace,
					applicationName+"-"+volumeName),
			},
			expectedResult:       reconcile.Result{},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName, statusPodFrozen),
		},
		{
			name: "no phase - should not freeze if things are missing",
			objs: []runtime.Object{
				testHelpers.NewPod(namespace, applicationName, 1),
				testHelpers.NewPersistentVolume(volumeName),
				testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
			},
			snapshotObjs:         []runtime.Object{},
			volumeBackup:         testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName, statusEmpty),
			expectedActions:      []core.Action{},
			expectedResult:       reconcile.Result{},
			expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName, statusEmpty),
			expectedError: errors.NewNotFound(schema.GroupResource{
				Group:    "apps",
				Resource: "deployments",
			}, applicationName),
		},
		// {
		// 	name: "reconcile add - VolumeBackup object has not been created yet",
		// 	objs: []runtime.Object{
		// 		testHelpers.NewDeployment(namespace, applicationName, &replicas),
		// 		testHelpers.NewPod(namespace, applicationName, 1),
		// 		testHelpers.NewPersistentVolume(volumeName),
		// 		testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
		// 	},
		// 	snapshotObjs:         []runtime.Object{},
		// 	expectedActions:      []core.Action{},
		// 	expectedVolumeBackup: nil,
		// 	expectedResult: reconcile.Result{
		// 		Requeue:      false,
		// 		RequeueAfter: 0,
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name:            "reconcile add - can't find deployment",
		// 	objs:            []runtime.Object{},
		// 	snapshotObjs:    []runtime.Object{},
		// 	volumeBackup:    testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName),
		// 	expectedActions: []core.Action{},
		// 	expectedResult:  reconcile.Result{},
		// 	expectedError: errors.NewNotFound(schema.GroupResource{
		// 		Group:    "apps",
		// 		Resource: "deployments",
		// 	}, applicationName),
		// 	expectedVolumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName),
		// },
		// {
		// 	name: "reconcile add - no pods available - basically a no-op",
		// 	objs: []runtime.Object{
		// 		testHelpers.NewDeployment(namespace, applicationName, &replicas),
		// 		testHelpers.NewPersistentVolume(volumeName),
		// 		testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
		// 	},
		// 	snapshotObjs:    []runtime.Object{},
		// 	volumeBackup:    testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName),
		// 	expectedActions: []core.Action{},
		// 	expectedResult:  reconcile.Result{Requeue: true},
		// },
		// {
		// 	name: "No phase - reconcile new add - volume does not exist",
		// 	objs: []runtime.Object{
		// 		testHelpers.NewDeployment(namespace, applicationName, &replicas),
		// 		testHelpers.NewPod(namespace, applicationName, 1),
		// 		testHelpers.NewPersistentVolumeClaim(namespace, claimName, volumeName),
		// 	},
		// 	snapshotObjs: []runtime.Object{},
		// 	volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName),
		// 	expectedVolumeBackup: &backupsv1alpha1.VolumeBackup{
		// 		TypeMeta: v1.TypeMeta{
		// 			Kind:       "VolumeBackup",
		// 			APIVersion: "v1alpha1",
		// 		},
		// 		ObjectMeta: v1.ObjectMeta{
		// 			Name:      name,
		// 			Namespace: namespace,
		// 		},
		// 		Spec: backupsv1alpha1.VolumeBackupSpec{
		// 			ApplicationName: applicationName,
		// 			VolumeName:      volumeName,
		// 			ContainerName:   containerName,
		// 		},
		// 		Status: backupsv1alpha1.VolumeBackupStatus{
		// 			VolumeBackupConditions: []backupsv1alpha1.VolumeBackupCondition{
		// 				{
		// 					Type:   backupsv1alpha1.PodFrozen,
		// 					Status: backupsv1alpha1.ConditionTrue,
		// 				},
		// 			},
		// 		},
		// 	},
		// 	expectedActions: []core.Action{
		// 		core.NewGetAction(gvr,
		// 			namespace,
		// 			applicationName+"-"+volumeName),
		// 	},
		// 	expectedError: nil,
		// },
		// {
		// 	name: "reconcile add - volume claim does not exist",
		// 	objs: []runtime.Object{
		// 		testHelpers.NewDeployment(namespace, applicationName, &replicas),
		// 		testHelpers.NewPod(namespace, applicationName, 1),
		// 		testHelpers.NewPersistentVolume(volumeName),
		// 	},
		// 	snapshotObjs: []runtime.Object{},
		// 	volumeBackup: testHelpers.NewVolumeBackup(namespace, name, applicationName, containerName, volumeName),
		// 	expectedActions: []core.Action{
		// 		core.NewGetAction(gvr, namespace, applicationName+"-"+volumeName),
		// 	},
		// 	expectedVolumeBackup: &backupsv1alpha1.VolumeBackup{
		// 		TypeMeta: v1.TypeMeta{
		// 			Kind:       "VolumeBackup",
		// 			APIVersion: "v1alpha1",
		// 		},
		// 		ObjectMeta: v1.ObjectMeta{
		// 			Name:      name,
		// 			Namespace: namespace,
		// 		},
		// 		Spec: backupsv1alpha1.VolumeBackupSpec{
		// 			ApplicationName: applicationName,
		// 			VolumeName:      volumeName,
		// 			ContainerName:   containerName,
		// 		},
		// 		Status: backupsv1alpha1.VolumeBackupStatus{
		// 			VolumeBackupConditions: []backupsv1alpha1.VolumeBackupCondition{
		// 				{
		// 					Type:   backupsv1alpha1.PodFrozen,
		// 					Status: backupsv1alpha1.ConditionTrue,
		// 				},
		// 			},
		// 		},
		// 	},
		// 	expectedError: nil,
		// },
	}

	for _, testCase := range cases {
		if !testCase.skip {
			runInTestHarness(t, testCase)
		}
	}
}
