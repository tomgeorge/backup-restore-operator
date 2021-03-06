package helperstest

import (
	"fmt"
	"reflect"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	volumebackupv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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
		Group:    "snapshot.storage.k8s.io",
		Resource: "volumesnapshots",
		Version:  "v1alpha1",
	}
	StatusEmpty = &volumebackupv1alpha1.VolumeBackupStatus{}
)

type CreateOpts struct {
	Namespace          string
	VolumeBackupName   string
	ApplicationName    string
	ContainerName      string
	VolumeName         string
	VolumeBackupStatus *volumebackupv1alpha1.VolumeBackupStatus
}

func NewVolumeBackup(namespace, volumeBackupName, applicationName, containerName, volumeName string, status *volumebackupv1alpha1.VolumeBackupStatus) *volumebackupv1alpha1.VolumeBackup {
	return &volumebackupv1alpha1.VolumeBackup{
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

func NewReadyVolumeSnapshot(namespace, snapshotName, claimName string) *v1alpha1.VolumeSnapshot {
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
		Status: v1alpha1.VolumeSnapshotStatus{
			ReadyToUse:   true,
			CreationTime: &v1.Time{},
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
			Name: fmt.Sprintf("pod-named-volume-%d", i),
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

// This is an awful hack where I copied
// github.com/kubernetes-csi/external-snapshotter/cmd/csi-snapshotter/main.CreateCRDs
// and splatted it here so I can get the envtest server to understand these kinds without having access
// to CRD yamls for them
func CreateSnapshotCRDs() []*apiextensionsv1beta1.CustomResourceDefinition {
	return []*apiextensionsv1beta1.CustomResourceDefinition{
		&apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: v1alpha1.VolumeSnapshotClassResourcePlural + "." + v1alpha1.GroupName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   v1alpha1.GroupName,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Scope:   apiextensionsv1beta1.ClusterScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural: v1alpha1.VolumeSnapshotClassResourcePlural,
					Kind:   reflect.TypeOf(v1alpha1.VolumeSnapshotClass{}).Name(),
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
		&apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: v1alpha1.VolumeSnapshotContentResourcePlural + "." + v1alpha1.GroupName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   v1alpha1.GroupName,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Scope:   apiextensionsv1beta1.ClusterScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural: v1alpha1.VolumeSnapshotContentResourcePlural,
					Kind:   reflect.TypeOf(v1alpha1.VolumeSnapshotContent{}).Name(),
				},
			},
		},
		&apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: v1alpha1.VolumeSnapshotResourcePlural + "." + v1alpha1.GroupName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   v1alpha1.GroupName,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural: v1alpha1.VolumeSnapshotResourcePlural,
					Kind:   reflect.TypeOf(v1alpha1.VolumeSnapshot{}).Name(),
				},
				Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
					Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
				},
			},
		},
	}
}
