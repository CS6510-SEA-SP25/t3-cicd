/*
Copyright 2025.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SecretReference struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

type RabbitMQConfig struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=5672
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// Reference to a Kubernetes Secret containing the password
	// +kubebuilder:validation:Required
	PasswordSecretRef SecretReference `json:"passwordSecretRef"`

	// +kubebuilder:validation:Required
	QueueName string `json:"queueName"`
}

type DatabaseConfig struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=3306
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// +kubebuilder:validation:Required
	PasswordSecretRef SecretReference `json:"passwordSecretRef"`

	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// +kubebuilder:default="false"
	// +kubebuilder:validation:Enum="true";"false"
	SSLMode string `json:"sslMode,omitempty"`

	// Reference to a Kubernetes Secret containing the SSL CA certificate
	// +optional
	SSLCASecretRef *SecretReference `json:"sslCASecretRef,omitempty"`
}

type StorageConfig struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Required
	AccessKeyRef SecretReference `json:"accessKeyRef"`

	// +kubebuilder:validation:Required
	SecretKeyRef SecretReference `json:"secretKeyRef"`

	// +kubebuilder:validation:Required
	DefaultBucket string `json:"defaultBucket"`
}

type CacheConfig struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=3306
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// +kubebuilder:validation:Required
	PasswordSecretRef SecretReference `json:"passwordSecretRef"`
}

type WorkerPoolConfig struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=5
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// +kubebuilder:validation:Required
	WorkerImage string `json:"workerImage"`

	// +kubebuilder:default="latest"
	WorkerImageTag string `json:"workerImageTag,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type ScalingConfig struct {
	// // +kubebuilder:validation:Minimum=1
	// // +kubebuilder:default=1
	// MessagesPerWorker int32 `json:"messagesPerWorker,omitempty"`

	// // +kubebuilder:validation:Minimum=1
	// // +kubebuilder:default=30
	// ScaleUpThreshold int32 `json:"scaleUpThreshold,omitempty"`

	// // +kubebuilder:validation:Minimum=1
	// // +kubebuilder:default=5
	// ScaleDownThreshold int32 `json:"scaleDownThreshold,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=30
	PollingIntervalSeconds int32 `json:"pollingIntervalSeconds,omitempty"`

	// // +kubebuilder:default=300
	// CooldownPeriodSeconds int32 `json:"cooldownPeriodSeconds,omitempty"`
}

// PoolScalerSpec defines the desired state of PoolScaler.
type PoolScalerSpec struct {
	// +kubebuilder:validation:Required
	InputQueue RabbitMQConfig `json:"inputQueue"`

	// +kubebuilder:validation:Optional
	OutputQueue RabbitMQConfig `json:"outputQueue,omitempty"`

	// +kubebuilder:validation:Required
	Database DatabaseConfig `json:"database,omitempty"`

	// +kubebuilder:validation:Required
	Storage StorageConfig `json:"storage,omitempty"`

	// +kubebuilder:validation:Required
	Cache CacheConfig `json:"cache,omitempty"`

	// +kubebuilder:validation:Required
	WorkerPool WorkerPoolConfig `json:"workerPool"`

	// +kubebuilder:validation:Required
	Scaling ScalingConfig `json:"scaling"`
}

// PoolScalerStatus defines the observed state of PoolScaler.
type PoolScalerStatus struct {
	// +kubebuilder:validation:Enum=Pending;Running;Error
	Phase string `json:"phase,omitempty"`

	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	QueueMessages int32 `json:"queueMessages,omitempty"`

	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.currentReplicas"
// +kubebuilder:printcolumn:name="Messages",type="integer",JSONPath=".status.queueMessages"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PoolScaler is the Schema for the poolscalers API.
type PoolScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PoolScalerSpec   `json:"spec,omitempty"`
	Status PoolScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PoolScalerList contains a list of PoolScaler.
type PoolScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PoolScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PoolScaler{}, &PoolScalerList{})
}
