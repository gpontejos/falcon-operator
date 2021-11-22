// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconAPI) DeepCopyInto(out *FalconAPI) {
	*out = *in
	if in.CID != nil {
		in, out := &in.CID, &out.CID
		*out = new(string)
		**out = **in
	}
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
	in.Spec.DeepCopyInto(&out.Spec)
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
	in.FalconAPI.DeepCopyInto(&out.FalconAPI)
	in.Registry.DeepCopyInto(&out.Registry)
	if in.InstallerArgs != nil {
		in, out := &in.InstallerArgs, &out.InstallerArgs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
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
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
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
func (in *FalconNodeSensor) DeepCopyInto(out *FalconNodeSensor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconNodeSensor.
func (in *FalconNodeSensor) DeepCopy() *FalconNodeSensor {
	if in == nil {
		return nil
	}
	out := new(FalconNodeSensor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FalconNodeSensor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconNodeSensorConfig) DeepCopyInto(out *FalconNodeSensorConfig) {
	*out = *in
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconNodeSensorConfig.
func (in *FalconNodeSensorConfig) DeepCopy() *FalconNodeSensorConfig {
	if in == nil {
		return nil
	}
	out := new(FalconNodeSensorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconNodeSensorList) DeepCopyInto(out *FalconNodeSensorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FalconNodeSensor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconNodeSensorList.
func (in *FalconNodeSensorList) DeepCopy() *FalconNodeSensorList {
	if in == nil {
		return nil
	}
	out := new(FalconNodeSensorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FalconNodeSensorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconNodeSensorSpec) DeepCopyInto(out *FalconNodeSensorSpec) {
	*out = *in
	in.Node.DeepCopyInto(&out.Node)
	in.Falcon.DeepCopyInto(&out.Falcon)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconNodeSensorSpec.
func (in *FalconNodeSensorSpec) DeepCopy() *FalconNodeSensorSpec {
	if in == nil {
		return nil
	}
	out := new(FalconNodeSensorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconNodeSensorStatus) DeepCopyInto(out *FalconNodeSensorStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconNodeSensorStatus.
func (in *FalconNodeSensorStatus) DeepCopy() *FalconNodeSensorStatus {
	if in == nil {
		return nil
	}
	out := new(FalconNodeSensorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FalconSensor) DeepCopyInto(out *FalconSensor) {
	*out = *in
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FalconSensor.
func (in *FalconSensor) DeepCopy() *FalconSensor {
	if in == nil {
		return nil
	}
	out := new(FalconSensor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistrySpec) DeepCopyInto(out *RegistrySpec) {
	*out = *in
	out.TLS = in.TLS
	if in.AcrName != nil {
		in, out := &in.AcrName, &out.AcrName
		*out = new(string)
		**out = **in
	}
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
