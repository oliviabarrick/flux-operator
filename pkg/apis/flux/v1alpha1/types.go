package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FluxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Flux `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Flux struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FluxSpec   `json:"spec"`
	Status            FluxStatus `json:"status,omitempty"`
}

type FluxSpec struct {
	GitUrl          string `json:"gitUrl,omitempty"`
	GitBranch       string `json:"gitBranch,omitempty"`
	GitPath         string `json:"gitPath,omitempty"`
	GitPollInterval string `json:"gitPollInterval,omitempty"`
	GitSecret       string `json:"gitSecret,omitempty"`
	FluxImage       string `json:"fluxImage,omitempty"`
	FluxVersion     string `json:"fluxVersion,omitempty"`
	Args            map[string]string `json:"args"`
	Role            FluxRole `json:"role"`
	ClusterRole     FluxRole `json:"clusterRole"`
}

type FluxRole struct {
	Enabled bool `json:"enabled"`
	Rules   []rbacv1.PolicyRule `json:"rules,omitempty"`
}

type FluxStatus struct {
	// Fill me
}
