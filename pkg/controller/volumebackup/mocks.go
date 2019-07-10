package volumebackup

import (
	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newVolumeBackup(namespace, volumeBackupName, applicationRef string) *volumebackupv1alpha1.VolumeBackup {
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
			ApplicationRef: applicationRef,
		},
	}
}

func newVolumeBackupList() *volumebackupv1alpha1.VolumeBackupList {
	return &volumebackupv1alpha1.VolumeBackupList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: volumebackupv1alpha1.SchemeGroupVersion.String(),
			Kind:       "VolumeBackupList",
		},
		ListMeta: metav1.ListMeta{},
		Items:    []volumebackupv1alpha1.VolumeBackup{},
	}
}

func newDeployment(namespace, name string, replicas *int32) *appsv1.Deployment {
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

func newPod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    name,
				Image:   "busybox",
				Command: []string{"echo hello"},
			}},
			Volumes: []corev1.Volume{{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "test-claim",
					},
				},
			}},
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
