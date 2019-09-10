// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackup) DeepCopyInto(out *VolumeBackup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackup.
func (in *VolumeBackup) DeepCopy() *VolumeBackup {
	if in == nil {
		return nil
	}
	out := new(VolumeBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeBackup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupCondition) DeepCopyInto(out *VolumeBackupCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupCondition.
func (in *VolumeBackupCondition) DeepCopy() *VolumeBackupCondition {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupList) DeepCopyInto(out *VolumeBackupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeBackup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupList.
func (in *VolumeBackupList) DeepCopy() *VolumeBackupList {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeBackupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupProvider) DeepCopyInto(out *VolumeBackupProvider) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupProvider.
func (in *VolumeBackupProvider) DeepCopy() *VolumeBackupProvider {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeBackupProvider) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupProviderList) DeepCopyInto(out *VolumeBackupProviderList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumeBackupProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupProviderList.
func (in *VolumeBackupProviderList) DeepCopy() *VolumeBackupProviderList {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupProviderList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeBackupProviderList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupProviderSpec) DeepCopyInto(out *VolumeBackupProviderSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupProviderSpec.
func (in *VolumeBackupProviderSpec) DeepCopy() *VolumeBackupProviderSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupProviderSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupProviderStatus) DeepCopyInto(out *VolumeBackupProviderStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupProviderStatus.
func (in *VolumeBackupProviderStatus) DeepCopy() *VolumeBackupProviderStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupProviderStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupSpec) DeepCopyInto(out *VolumeBackupSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupSpec.
func (in *VolumeBackupSpec) DeepCopy() *VolumeBackupSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeBackupStatus) DeepCopyInto(out *VolumeBackupStatus) {
	*out = *in
	if in.VolumeBackupConditions != nil {
		in, out := &in.VolumeBackupConditions, &out.VolumeBackupConditions
		*out = make([]VolumeBackupCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeBackupStatus.
func (in *VolumeBackupStatus) DeepCopy() *VolumeBackupStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeBackupStatus)
	in.DeepCopyInto(out)
	return out
}
