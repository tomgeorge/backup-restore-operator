package util

import (
	"fmt"

	v1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackupRestoreUtils struct {
	Client client.Client
	Cfg    *rest.Config
}

func (u *BackupRestoreUtils) GetVolumeSnapshotFromCR(cr *v1alpha1.VolumeBackup, pod *corev1.Pod) string {
	volumeName := cr.Spec.VolumeName
	persistentVolumeClaimName := ""

	for _, vol := range pod.Spec.Volumes {
		if vol.VolumeSource.PersistentVolumeClaim != nil && vol.VolumeSource.PersistentVolumeClaim.ClaimName == volumeName {
			persistentVolumeClaimName = vol.VolumeSource.PersistentVolumeClaim.ClaimName
		}
	}
	return fmt.Sprintf("%s-%s", cr.Spec.ApplicationName, persistentVolumeClaimName)
}
