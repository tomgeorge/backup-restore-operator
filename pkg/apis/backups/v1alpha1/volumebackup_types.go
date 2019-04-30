package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolumeBackupSpec defines the desired state of VolumeBackup
type VolumeBackupSpec struct {
	volumeBackupProviderRef string `json:"volumeBackupProviderRef"`
	volumeClaimRef          string `json:"volumeClaimRef"`
	authenticationSecret    string `json:"authenticationSecret"`
	ApplicationRef          string `json:"applicationRef"`
}

// VolumeBackupStatus defines the observed state of VolumeBackup
type VolumeBackupStatus struct {
	backupRequests []VolumeBackup `json:"backupRequests"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBackup is the Schema for the volumebackups API
// +k8s:openapi-gen=true
type VolumeBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeBackupSpec   `json:"spec,omitempty"`
	Status VolumeBackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBackupList contains a list of VolumeBackup
type VolumeBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeBackup{}, &VolumeBackupList{})
}
