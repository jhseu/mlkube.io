// +build !ignore_autogenerated

// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AcceleratorConfig) DeepCopyInto(out *AcceleratorConfig) {
	*out = *in
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]AcceleratorVolume, len(*in))
		copy(*out, *in)
	}
	if in.EnvVars != nil {
		in, out := &in.EnvVars, &out.EnvVars
		*out = make([]EnvironmentVariableConfig, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AcceleratorConfig.
func (in *AcceleratorConfig) DeepCopy() *AcceleratorConfig {
	if in == nil {
		return nil
	}
	out := new(AcceleratorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AcceleratorVolume) DeepCopyInto(out *AcceleratorVolume) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AcceleratorVolume.
func (in *AcceleratorVolume) DeepCopy() *AcceleratorVolume {
	if in == nil {
		return nil
	}
	out := new(AcceleratorVolume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChiefSpec) DeepCopyInto(out *ChiefSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChiefSpec.
func (in *ChiefSpec) DeepCopy() *ChiefSpec {
	if in == nil {
		return nil
	}
	out := new(ChiefSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerConfig) DeepCopyInto(out *ControllerConfig) {
	*out = *in
	if in.Accelerators != nil {
		in, out := &in.Accelerators, &out.Accelerators
		*out = make(map[string]AcceleratorConfig, len(*in))
		for key, val := range *in {
			newVal := new(AcceleratorConfig)
			val.DeepCopyInto(newVal)
			(*out)[key] = *newVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerConfig.
func (in *ControllerConfig) DeepCopy() *ControllerConfig {
	if in == nil {
		return nil
	}
	out := new(ControllerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvironmentVariableConfig) DeepCopyInto(out *EnvironmentVariableConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvironmentVariableConfig.
func (in *EnvironmentVariableConfig) DeepCopy() *EnvironmentVariableConfig {
	if in == nil {
		return nil
	}
	out := new(EnvironmentVariableConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFJob) DeepCopyInto(out *TFJob) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFJob.
func (in *TFJob) DeepCopy() *TFJob {
	if in == nil {
		return nil
	}
	out := new(TFJob)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TFJob) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFJobList) DeepCopyInto(out *TFJobList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TFJob, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFJobList.
func (in *TFJobList) DeepCopy() *TFJobList {
	if in == nil {
		return nil
	}
	out := new(TFJobList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TFJobList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFJobSpec) DeepCopyInto(out *TFJobSpec) {
	*out = *in
	if in.ReplicaSpecs != nil {
		in, out := &in.ReplicaSpecs, &out.ReplicaSpecs
		*out = make([]*TFReplicaSpec, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} else {
				(*out)[i] = new(TFReplicaSpec)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	if in.TerminationPolicy != nil {
		in, out := &in.TerminationPolicy, &out.TerminationPolicy
		if *in == nil {
			*out = nil
		} else {
			*out = new(TerminationPolicySpec)
			(*in).DeepCopyInto(*out)
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFJobSpec.
func (in *TFJobSpec) DeepCopy() *TFJobSpec {
	if in == nil {
		return nil
	}
	out := new(TFJobSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFJobStatus) DeepCopyInto(out *TFJobStatus) {
	*out = *in
	if in.ReplicaStatuses != nil {
		in, out := &in.ReplicaStatuses, &out.ReplicaStatuses
		*out = make([]*TFReplicaStatus, len(*in))
		for i := range *in {
			if (*in)[i] == nil {
				(*out)[i] = nil
			} else {
				(*out)[i] = new(TFReplicaStatus)
				(*in)[i].DeepCopyInto((*out)[i])
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFJobStatus.
func (in *TFJobStatus) DeepCopy() *TFJobStatus {
	if in == nil {
		return nil
	}
	out := new(TFJobStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFReplicaSpec) DeepCopyInto(out *TFReplicaSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		if *in == nil {
			*out = nil
		} else {
			*out = new(int32)
			**out = **in
		}
	}
	if in.Template != nil {
		in, out := &in.Template, &out.Template
		if *in == nil {
			*out = nil
		} else {
			*out = new(v1.PodTemplateSpec)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.TFPort != nil {
		in, out := &in.TFPort, &out.TFPort
		if *in == nil {
			*out = nil
		} else {
			*out = new(int32)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFReplicaSpec.
func (in *TFReplicaSpec) DeepCopy() *TFReplicaSpec {
	if in == nil {
		return nil
	}
	out := new(TFReplicaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TFReplicaStatus) DeepCopyInto(out *TFReplicaStatus) {
	*out = *in
	if in.ReplicasStates != nil {
		in, out := &in.ReplicasStates, &out.ReplicasStates
		*out = make(map[ReplicaState]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TFReplicaStatus.
func (in *TFReplicaStatus) DeepCopy() *TFReplicaStatus {
	if in == nil {
		return nil
	}
	out := new(TFReplicaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TerminationPolicySpec) DeepCopyInto(out *TerminationPolicySpec) {
	*out = *in
	if in.Chief != nil {
		in, out := &in.Chief, &out.Chief
		if *in == nil {
			*out = nil
		} else {
			*out = new(ChiefSpec)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TerminationPolicySpec.
func (in *TerminationPolicySpec) DeepCopy() *TerminationPolicySpec {
	if in == nil {
		return nil
	}
	out := new(TerminationPolicySpec)
	in.DeepCopyInto(out)
	return out
}
