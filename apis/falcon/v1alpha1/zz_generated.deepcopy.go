// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconAPI) DeepCopyInto(out *FalconAPI) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconAPI.
func (in *FalconAPI) DeepCopy() *FalconAPI {
	if in == nil {
		return nil
	}
	out := new(FalconAPI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconContainer) DeepCopyInto(out *FalconContainer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconContainer.
func (in *FalconContainer) DeepCopy() *FalconContainer {
	if in == nil {
		return nil
	}
	out := new(FalconContainer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FalconContainer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconContainerList) DeepCopyInto(out *FalconContainerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FalconContainer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconContainerList.
func (in *FalconContainerList) DeepCopy() *FalconContainerList {
	if in == nil {
		return nil
	}
	out := new(FalconContainerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FalconContainerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconContainerSpec) DeepCopyInto(out *FalconContainerSpec) {
	*out = *in
	out.FalconAPI = in.FalconAPI
	out.Registry = in.Registry
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconContainerSpec.
func (in *FalconContainerSpec) DeepCopy() *FalconContainerSpec {
	if in == nil {
		return nil
	}
	out := new(FalconContainerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconContainerStatus) DeepCopyInto(out *FalconContainerStatus) {
	*out = *in
	if in.RetryAttempt != nil {
		in, out := &in.RetryAttempt, &out.RetryAttempt
		*out = new(uint8)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconContainerStatus.
func (in *FalconContainerStatus) DeepCopy() *FalconContainerStatus {
	if in == nil {
		return nil
	}
	out := new(FalconContainerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistrySpec) DeepCopyInto(out *RegistrySpec) {
	*out = *in
	out.TLS = in.TLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistrySpec.
func (in *RegistrySpec) DeepCopy() *RegistrySpec {
	if in == nil {
		return nil
	}
	out := new(RegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryTLSSpec) DeepCopyInto(out *RegistryTLSSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryTLSSpec.
func (in *RegistryTLSSpec) DeepCopy() *RegistryTLSSpec {
	if in == nil {
		return nil
	}
	out := new(RegistryTLSSpec)
	in.DeepCopyInto(out)
	return out
}