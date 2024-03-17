/*
Copyright 2024 nineinfra.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceConfig struct {
	// The replicas of the cluster workload.Default value is 1
	// +optional
	Replicas int32 `json:"replicas"`
	// num of the disks. default value is 1
	// +optional
	Disks int32 `json:"disks"`
	// the storage class. default value is nineinfra-default
	// +optional
	StorageClass string `json:"storageClass"`
	// The resource requirements of the cluster workload.
	// +optional
	ResourceRequirements corev1.ResourceRequirements `json:"resourceRequirements"`
}

type ImageConfig struct {
	Repository string `json:"repository"`
	// Image tag. Usually the vesion of the cluster, default: `latest`.
	// +optional
	Tag string `json:"tag,omitempty"`
	// Image pull policy. One of `Always, Never, IfNotPresent`, default: `Always`.
	// +kubebuilder:default:=Always
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	PullPolicy string `json:"pullPolicy,omitempty"`
	// Secrets for image pull.
	// +optional
	PullSecrets string `json:"pullSecret,omitempty"`
}

type SchemeConfig struct {
	Fields map[string]string `json:"fields"`
}

type SourceConfig struct {
	// Type. The type of the source.
	Type string `json:"type"`
	// Conf. The conf of the source.
	Conf map[string]string `json:"conf"`
	// TableList. The table list of the source.
	// +optional
	TableList []map[string]string `json:"table_list"`
	// Scheme.  The scheme of the source.
	// +optional
	Scheme SchemeConfig `json:"scheme"`
	// Properties.  The scheme of the source.
	// +optional
	Properties map[string]string `json:"properties"`
	// ExtraConfig. The extra config k/v of the sink.
	ExtraConfig map[string]string `json:"extra_config"`
}

type TransformConfig struct {
	// Type. The type of the transform.
	Type string `json:"type"`
	// Conf. The conf of the transform.
	Conf map[string]string `json:"conf"`
	// Scheme.  The scheme of the source.
	// +optional
	Scheme SchemeConfig `json:"scheme"`
}

type SinkConfig struct {
	// Type. The type of the sink.
	Type string `json:"type"`
	// Conf. The conf of the sink.
	Conf map[string]string `json:"conf"`
	// PartitionBy. The partition field list of the sink.
	PartitionBy []string `json:"partition_by"`
	// PartitionBy. The sink column list of the sink.
	SinkColumns []string `json:"sink_columns"`
	// ExtraConfig. The extra config k/v of the sink.
	ExtraConfig map[string]string `json:"extra_config"`
}

type SeatunnelConfig struct {
	// Env. k/v configs for the env section of the seatunnel.conf.
	// +optional
	Env map[string]string `json:"env"`
	// Source. k/v configs for the source section of the seatunnel.conf.
	Source SourceConfig `json:"source"`
	// Source. k/v configs for the transform section of the seatunnel.conf.
	// +optional
	Transform TransformConfig `json:"transform"`
	// Source. k/v configs for the sink section of the seatunnel.conf.
	Sink SinkConfig `json:"sink"`
}

// SeatunnelJobSpec defines the desired state of SeatunnelJob
type SeatunnelJobSpec struct {
	// Version. version of the job.
	Version string `json:"version"`
	// Image. image config of the job.
	Image ImageConfig `json:"image"`
	// Resource. resouce config of the job.
	// +optional
	Resource ResourceConfig `json:"resource,omitempty"`
	// Conf. conf of the seatunnel job
	Conf SeatunnelConfig `json:"conf"`
	// K8sConf. k/v configs for the cluster in k8s.such as the cluster domain
	// +optional
	K8sConf map[string]string `json:"k8sConf,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeatunnelJob is the Schema for the seatunneljobs API
type SeatunnelJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeatunnelJobSpec   `json:"spec,omitempty"`
	Status SeatunnelJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeatunnelJobList contains a list of SeatunnelJob
type SeatunnelJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeatunnelJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeatunnelJob{}, &SeatunnelJobList{})
}
