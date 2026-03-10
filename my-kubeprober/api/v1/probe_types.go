package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProbeSpec defines the desired state of Probe
type ProbeSpec struct {
	// Policy 定义了探测策略，比如运行频率
	Policy ProbePolicy `json:"policy,omitempty"`

	// Template 定义了要运行的 Pod 模板
	// 我们直接复用 K8s 原生的 PodTemplateSpec，这样用户可以使用所有 Pod 的特性
	// +kubebuilder:pruning:PreserveUnknownFields
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
}

type ProbePolicy struct {
	// RunInterval 定义了探测任务的运行间隔（秒）
	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=1
	RunInterval int32 `json:"runInterval,omitempty"`
}

// ProbeStatus defines the observed state of Probe
type ProbeStatus struct {
	// Phase 记录当前探测任务的整体状态 (Running, Pending, Succeeded, Failed)
	Phase string `json:"phase,omitempty"`

	// LastRun 记录上一次任务开始的时间
	LastRun *metav1.Time `json:"lastRun,omitempty"`

	// LastResult 记录上一次探测的业务结果 (Succeeded / Failed)
	LastResult string `json:"lastResult,omitempty"`

	// LastMessage 记录探测失败时的简短原因 (例如 "ExitCode: 1")
	LastMessage string `json:"lastMessage,omitempty"`

	// TotalRuns 记录累计运行次数
	TotalRuns int32 `json:"totalRuns,omitempty"`

	// Conditions 记录详细的状态变迁历史
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="The current phase of the probe"
//+kubebuilder:printcolumn:name="LastResult",type="string",JSONPath=".status.lastResult",description="The result of the last probe run"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Probe is the Schema for the probes API
type Probe struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProbeSpec   `json:"spec,omitempty"`
	Status ProbeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProbeList contains a list of Probe
type ProbeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Probe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Probe{}, &ProbeList{})
}
