package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolumeBackupSpec defines the desired state of VolumeBackup
// +k8s:openapi-gen=true
type VolumeBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	ApplicationName string `json:"applicationName"`
	VolumeName      string `json:"volumeName"`
	ContainerName   string `json:"containerName"`
}

// VolumeBackupStatus defines the observed state of VolumeBackup
type VolumeBackupStatus struct {
	// The list of VolumeBackupConditions that the Backup goes through
	Conditions []VolumeBackupCondition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VolumeBackupStatusList contains a list of VolumeBackup
// type VolumeBackupStatusList struct {
// 	metav1.TypeMeta `json:",inline"`
// 	metav1.ListMeta `json:"metadata,omitempty"`
// 	Items           []VolumeBackupStatus `json:"items"`
// }

// // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// // VolumeBackupConditionList contains a list of VolumeBackup
// type VolumeBackupConditionList struct {
// 	metav1.TypeMeta `json:",inline"`
// 	metav1.ListMeta `json:"metadata,omitempty"`
// 	Items           []VolumeBackupCondition `json:"items"`
// }

type VolumeBackupCondition struct {
	// Type is the type of the condition.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions
	Type VolumeBackupConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions
	// +optional
	Status ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	// LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

///
// Status:
//   Conditions:
//   [
//     [ PodFrozen, True]
//     [ SnapshotIssued, True]
//     [ SnapshotCreated, False]?
const (
	PodFrozen         VolumeBackupConditionType = "PodFrozen"
	SnapshotIssued    VolumeBackupConditionType = "SnapshotIssued"
	SnapshotCreated   VolumeBackupConditionType = "SnapshotCreated"
	PodUnfrozen       VolumeBackupConditionType = "PodUnfrozen"
	SnapshotReady     VolumeBackupConditionType = "SnapshotReady"
	SnapshotUploading VolumeBackupConditionType = "SnapshotUploading"
	SnapshotUploaded  VolumeBackupConditionType = "SnapshotUploaded"
)

// VolumeBackupConditionType is a valid value for VolumeBackupCondition.Type
type VolumeBackupConditionType string

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// VolumeBackup is the Schema for the volumebackups API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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

func (vb *VolumeBackup) IsFrozen() bool {
	return vb.checkStatus(PodFrozen)
}

func (vb *VolumeBackup) IsSnapshotIssued() bool {
	return vb.checkStatus(SnapshotIssued)
}

func (vb *VolumeBackup) IsSnapshotCreated() bool {
	return vb.checkStatus(SnapshotCreated)
}

func (vb *VolumeBackup) checkStatus(conditionType VolumeBackupConditionType) bool {
	for _, condition := range vb.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == ConditionTrue
		}
	}
	return false
}

func (vb *VolumeBackup) UpdateStatus(conditionType VolumeBackupConditionType, conditionStatus ConditionStatus, reason, message string) {
	newCondition := VolumeBackupCondition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	for idx, condition := range vb.Status.Conditions {
		if condition.Type == conditionType {
			vb.Status.Conditions[idx] = newCondition
			return
		}
	}
	vb.Status.Conditions = append(vb.Status.Conditions, newCondition)
	return
}
