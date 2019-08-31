package helperstest

import (
	"fmt"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	name                  = "example-backup"
	namespace             = "example"
	applicationName       = "example-application-to-backup"
	storageclass          = "mock-csi-storageclass"
	replicas        int32 = 1
	apiVersion            = "backups.example.com/v1alpha1"
	kind                  = "VolumeBackup"
	gvr                   = schema.GroupVersionResource{
		Group:    "volumesnapshot",
		Resource: "volumesnapshots",
		Version:  "v1alpha1",
	}
)

func NewVolumeBackup(namespace, volumeBackupName, applicationName, containerName, volumeName string, status *volumebackupv1alpha1.VolumeBackupStatus) *volumebackupv1alpha1.VolumeBackup {
	return &volumebackupv1alpha1.VolumeBackup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: volumebackupv1alpha1.SchemeGroupVersion.String(),
			Kind:       "VolumeBackup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      volumeBackupName,
			Namespace: namespace,
		},
		Spec: volumebackupv1alpha1.VolumeBackupSpec{
			ApplicationName: applicationName,
			VolumeName:      volumeName,
			ContainerName:   containerName,
		},
		Status: *status,
	}
}

func NewVolumeSnapshot(namespace, snapshotName, claimName string) *v1alpha1.VolumeSnapshot {
	return &v1alpha1.VolumeSnapshot{
		ObjectMeta: v1.ObjectMeta{
			Name:      snapshotName,
			Namespace: "example",
			OwnerReferences: []v1.OwnerReference{
				NewOwnerReference(apiVersion, kind, name),
			},
		},
		Spec: v1alpha1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &storageclass,
			Source: &corev1.TypedLocalObjectReference{
				Kind: "PersistentVolumeClaim",
				Name: claimName,
			},
		},
	}

}

func NewVolumeBackupList() *volumebackupv1alpha1.VolumeBackupList {
	return &volumebackupv1alpha1.VolumeBackupList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: volumebackupv1alpha1.SchemeGroupVersion.String(),
			Kind:       "VolumeBackupList",
		},
		ListMeta: metav1.ListMeta{},
		Items:    []volumebackupv1alpha1.VolumeBackup{},
	}
}

func NewDeployment(namespace, name string, replicas *int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "busybox",
						Image:   "busybox",
						Command: []string{"echo hello"},
					}},
				},
			},
		},
	}
}

func NewPersistentVolume(volumeName string) *corev1.PersistentVolume {
	volume := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      volumeName,
			Namespace: "",
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: storageclass,
		},
	}
	return volume
}

func NewPersistentVolumeClaim(namespace, claimName, volumeName string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claimName,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: volumeName,
		},
	}
}

func NewPod(namespace, name string, volumeCount int) *corev1.Pod {
	volumes := []corev1.Volume{}
	for i := 0; i < volumeCount; i++ {
		volume := corev1.Volume{
			Name: fmt.Sprintf("data-%d", i),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("test-claim-%d", i),
				},
			},
		}
		volumes = append(volumes, volume)
	}
	return &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "busybox",
				Image:   "busybox",
				Command: []string{"echo hello"},
			}},
			Volumes: volumes,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "Deployment",
					Name:       name,
				},
			},
			Labels: map[string]string{
				"name": name,
			},
			Name:      name + "-pod",
			Namespace: namespace,
			Annotations: map[string]string{
				"backups.example.com.pre-hook":  "echo freeze",
				"backups.example.com.post-hook": "echo unfreeze",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "busybox",
					Ready: true,
				},
			},
		},
	}
}

func NewPodWithCommand(namespace, name, command string, volumeCount int) *corev1.Pod {
	volumes := []corev1.Volume{}
	for i := 0; i < volumeCount; i++ {
		volume := corev1.Volume{
			Name: fmt.Sprintf("data-%d", i),
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("test-claim-%d", i),
				},
			},
		}
		volumes = append(volumes, volume)
	}
	return &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "busybox",
				Image:   "busybox",
				Command: []string{command},
			}},
			Volumes: volumes,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: corev1.SchemeGroupVersion.String(),
					Kind:       "Deployment",
					Name:       name,
				},
			},
			Labels: map[string]string{
				"name": name,
			},
			Name:      name + "-pod",
			Namespace: namespace,
			Annotations: map[string]string{
				"backups.example.com.pre-hook":  "echo freeze",
				"backups.example.com.post-hook": "echo unfreeze",
			},
		},
	}
}

func NewOwnerReference(apiversion, kind, name string) v1.OwnerReference {
	isTrue := bool(true)
	return v1.OwnerReference{
		APIVersion:         apiversion,
		Kind:               kind,
		Name:               name,
		Controller:         &isTrue,
		BlockOwnerDeletion: &isTrue,
	}
}
