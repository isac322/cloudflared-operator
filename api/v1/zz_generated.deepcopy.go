//go:build !ignore_autogenerated

/*
Copyright 2024 Byeonghoon Yoo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Deployment) DeepCopyInto(out *Deployment) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	in.DeploymentStrategy.DeepCopyInto(&out.DeploymentStrategy)
	in.DaemonSetUpdateStrategy.DeepCopyInto(&out.DaemonSetUpdateStrategy)
	if in.RevisionHistoryLimit != nil {
		in, out := &in.RevisionHistoryLimit, &out.RevisionHistoryLimit
		*out = new(int32)
		**out = **in
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodLabels != nil {
		in, out := &in.PodLabels, &out.PodLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PodAnnotations != nil {
		in, out := &in.PodAnnotations, &out.PodAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Deployment.
func (in *Deployment) DeepCopy() *Deployment {
	if in == nil {
		return nil
	}
	out := new(Deployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginAccessSettings) DeepCopyInto(out *OriginAccessSettings) {
	*out = *in
	if in.Access != nil {
		in, out := &in.Access, &out.Access
		*out = new(OriginAccessSettingsAccess)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginAccessSettings.
func (in *OriginAccessSettings) DeepCopy() *OriginAccessSettings {
	if in == nil {
		return nil
	}
	out := new(OriginAccessSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginAccessSettingsAccess) DeepCopyInto(out *OriginAccessSettingsAccess) {
	*out = *in
	if in.Required != nil {
		in, out := &in.Required, &out.Required
		*out = new(bool)
		**out = **in
	}
	if in.TeamName != nil {
		in, out := &in.TeamName, &out.TeamName
		*out = new(string)
		**out = **in
	}
	if in.AudTag != nil {
		in, out := &in.AudTag, &out.AudTag
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginAccessSettingsAccess.
func (in *OriginAccessSettingsAccess) DeepCopy() *OriginAccessSettingsAccess {
	if in == nil {
		return nil
	}
	out := new(OriginAccessSettingsAccess)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginConfiguration) DeepCopyInto(out *OriginConfiguration) {
	*out = *in
	if in.TLSSettings != nil {
		in, out := &in.TLSSettings, &out.TLSSettings
		*out = new(OriginTLSSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTPSettings != nil {
		in, out := &in.HTTPSettings, &out.HTTPSettings
		*out = new(OriginHTTPSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.ConnectionSettings != nil {
		in, out := &in.ConnectionSettings, &out.ConnectionSettings
		*out = new(OriginConnectionSettings)
		(*in).DeepCopyInto(*out)
	}
	if in.AccessSettings != nil {
		in, out := &in.AccessSettings, &out.AccessSettings
		*out = new(OriginAccessSettings)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginConfiguration.
func (in *OriginConfiguration) DeepCopy() *OriginConfiguration {
	if in == nil {
		return nil
	}
	out := new(OriginConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginConnectionSettings) DeepCopyInto(out *OriginConnectionSettings) {
	*out = *in
	if in.ConnectTimeout != nil {
		in, out := &in.ConnectTimeout, &out.ConnectTimeout
		*out = new(string)
		**out = **in
	}
	if in.NoHappyEyeballs != nil {
		in, out := &in.NoHappyEyeballs, &out.NoHappyEyeballs
		*out = new(bool)
		**out = **in
	}
	if in.ProxyType != nil {
		in, out := &in.ProxyType, &out.ProxyType
		*out = new(string)
		**out = **in
	}
	if in.ProxyAddress != nil {
		in, out := &in.ProxyAddress, &out.ProxyAddress
		*out = new(string)
		**out = **in
	}
	if in.ProxyPort != nil {
		in, out := &in.ProxyPort, &out.ProxyPort
		*out = new(int)
		**out = **in
	}
	if in.KeepAliveTimeout != nil {
		in, out := &in.KeepAliveTimeout, &out.KeepAliveTimeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.KeepAliveConnections != nil {
		in, out := &in.KeepAliveConnections, &out.KeepAliveConnections
		*out = new(int)
		**out = **in
	}
	if in.TCPKeepAlive != nil {
		in, out := &in.TCPKeepAlive, &out.TCPKeepAlive
		*out = new(metav1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginConnectionSettings.
func (in *OriginConnectionSettings) DeepCopy() *OriginConnectionSettings {
	if in == nil {
		return nil
	}
	out := new(OriginConnectionSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginHTTPSettings) DeepCopyInto(out *OriginHTTPSettings) {
	*out = *in
	if in.HTTPHostHeader != nil {
		in, out := &in.HTTPHostHeader, &out.HTTPHostHeader
		*out = new(string)
		**out = **in
	}
	if in.DisableChunkedEncoding != nil {
		in, out := &in.DisableChunkedEncoding, &out.DisableChunkedEncoding
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginHTTPSettings.
func (in *OriginHTTPSettings) DeepCopy() *OriginHTTPSettings {
	if in == nil {
		return nil
	}
	out := new(OriginHTTPSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginTLSSettings) DeepCopyInto(out *OriginTLSSettings) {
	*out = *in
	if in.OriginServerName != nil {
		in, out := &in.OriginServerName, &out.OriginServerName
		*out = new(string)
		**out = **in
	}
	if in.CAPool != nil {
		in, out := &in.CAPool, &out.CAPool
		*out = new(string)
		**out = **in
	}
	if in.NoTLSVerify != nil {
		in, out := &in.NoTLSVerify, &out.NoTLSVerify
		*out = new(bool)
		**out = **in
	}
	if in.TLSTimeout != nil {
		in, out := &in.TLSTimeout, &out.TLSTimeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.HTTP2Origin != nil {
		in, out := &in.HTTP2Origin, &out.HTTP2Origin
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginTLSSettings.
func (in *OriginTLSSettings) DeepCopy() *OriginTLSSettings {
	if in == nil {
		return nil
	}
	out := new(OriginTLSSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretKeyRef) DeepCopyInto(out *SecretKeyRef) {
	*out = *in
	if in.Key != nil {
		in, out := &in.Key, &out.Key
		*out = new(string)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretKeyRef.
func (in *SecretKeyRef) DeepCopy() *SecretKeyRef {
	if in == nil {
		return nil
	}
	out := new(SecretKeyRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tunnel) DeepCopyInto(out *Tunnel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tunnel.
func (in *Tunnel) DeepCopy() *Tunnel {
	if in == nil {
		return nil
	}
	out := new(Tunnel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Tunnel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelList) DeepCopyInto(out *TunnelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Tunnel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelList.
func (in *TunnelList) DeepCopy() *TunnelList {
	if in == nil {
		return nil
	}
	out := new(TunnelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TunnelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelRunParameters) DeepCopyInto(out *TunnelRunParameters) {
	*out = *in
	if in.GracePeriod != nil {
		in, out := &in.GracePeriod, &out.GracePeriod
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.Logfile != nil {
		in, out := &in.Logfile, &out.Logfile
		*out = new(string)
		**out = **in
	}
	if in.Loglevel != nil {
		in, out := &in.Loglevel, &out.Loglevel
		*out = new(string)
		**out = **in
	}
	if in.Pidfile != nil {
		in, out := &in.Pidfile, &out.Pidfile
		*out = new(string)
		**out = **in
	}
	if in.Protocol != nil {
		in, out := &in.Protocol, &out.Protocol
		*out = new(string)
		**out = **in
	}
	if in.Region != nil {
		in, out := &in.Region, &out.Region
		*out = new(string)
		**out = **in
	}
	if in.Retries != nil {
		in, out := &in.Retries, &out.Retries
		*out = new(int)
		**out = **in
	}
	if in.Tag != nil {
		in, out := &in.Tag, &out.Tag
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelRunParameters.
func (in *TunnelRunParameters) DeepCopy() *TunnelRunParameters {
	if in == nil {
		return nil
	}
	out := new(TunnelRunParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelSpec) DeepCopyInto(out *TunnelSpec) {
	*out = *in
	in.APITokenSecretRef.DeepCopyInto(&out.APITokenSecretRef)
	if in.SecretName != nil {
		in, out := &in.SecretName, &out.SecretName
		*out = new(string)
		**out = **in
	}
	if in.ConfigMapName != nil {
		in, out := &in.ConfigMapName, &out.ConfigMapName
		*out = new(string)
		**out = **in
	}
	in.DaemonDeployment.DeepCopyInto(&out.DaemonDeployment)
	if in.OriginConfiguration != nil {
		in, out := &in.OriginConfiguration, &out.OriginConfiguration
		*out = new(OriginConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.TunnelRunParameters != nil {
		in, out := &in.TunnelRunParameters, &out.TunnelRunParameters
		*out = new(TunnelRunParameters)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelSpec.
func (in *TunnelSpec) DeepCopy() *TunnelSpec {
	if in == nil {
		return nil
	}
	out := new(TunnelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelStatus) DeepCopyInto(out *TunnelStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]TunnelStatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelStatus.
func (in *TunnelStatus) DeepCopy() *TunnelStatus {
	if in == nil {
		return nil
	}
	out := new(TunnelStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelStatusCondition) DeepCopyInto(out *TunnelStatusCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelStatusCondition.
func (in *TunnelStatusCondition) DeepCopy() *TunnelStatusCondition {
	if in == nil {
		return nil
	}
	out := new(TunnelStatusCondition)
	in.DeepCopyInto(out)
	return out
}