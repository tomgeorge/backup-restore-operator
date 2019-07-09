package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolumeBackupProviderSpec defines the desired state of VolumeBackupProvider
// +k8s:openapi-gen=true
type VolumeBackupProviderSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// VolumeBackupProviderStatus defines the observed state of VolumeBackupProvider
// +k8s:openapi-gen=true
type VolumeBackupProviderStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBackupProvider is the Schema for the volumebackupproviders API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type VolumeBackupProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeBackupProviderSpec   `json:"spec,omitempty"`
	Status VolumeBackupProviderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBackupProviderList contains a list of VolumeBackupProvider
type VolumeBackupProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VolumeBackupProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VolumeBackupProvider{}, &VolumeBackupProviderList{})
}
