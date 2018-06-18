package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FluxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Flux `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Flux struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FluxSpec   `json:"spec"`
	Status            FluxStatus `json:"status,omitempty"`
}

// Settings for operating Flux
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FluxSpec struct {
	// Namespace to deploy Flux and Tiller into.
	Namespace       string `json:"namespace"`
	// The URL to the Git repository (required).
	GitUrl          string `json:"gitUrl"`
	// The git branch to use (default: `master`).
	GitBranch       string `json:"gitBranch,omitempty"`
	// The path with in the git repository to look for YAML in (default: `.`)
	GitPath         string `json:"gitPath,omitempty"`
	// The frequency with which to fetch the git repository (default: `5m0s`).
	GitPollInterval string `json:"gitPollInterval,omitempty"`
	// The Kubernetes secret to use for cloning, if it does not exist it will
	// be generated (default: `flux-$name-git-deploy` or `$GIT_SECRET_NAME`).
	GitSecret       string `json:"gitSecret,omitempty"`
	// The image to use for flux (default: `quay.io/weaveworks/flux` or `$FLUX_IMAGE`).
	FluxImage       string `json:"fluxImage,omitempty"`
	// The version to use for flux (default: `1.4.0` or `$FLUX_VERSION`).
	FluxVersion     string `json:"fluxVersion,omitempty"`
	// A map of args to pass to flux without `--` prepended.
	Args            map[string]string `json:"args,omitempty"`
	// A role to add to the service account (default: none)
	Role            FluxRole `json:"role,omitempty"`
	// A cluster role to add to the service account (default: none)
	ClusterRole     FluxRole `json:"clusterRole,omitempty"`
	// The tiller settings.
	Tiller          Tiller   `json:"tiller,omitempty"`
	// The Helm Operator settings.
	HelmOperator    HelmOperator `json:"helmOperator,omitempty"`
}

// Represents a Role or ClusterRole for the Flux service account user.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type FluxRole struct {
  // If enabled, a role will be assigned to the service account (default: false)
	Enabled bool `json:"enabled,omitempty"`
  // the list of rbac rules to use (default: full access).
	Rules   []rbacv1.PolicyRule `json:"rules,omitempty"`
}

// Settings for operating Tiller alongside Flux.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type Tiller struct {
	// Whether or not to deploy a tiller instance in the same namespace (default: false).
	Enabled         bool `json:"enabled,omitempty"`
	// The image to use with tiller (default: `gcr.io/kubernetes-helm/tiller` or `$TILLER_IMAGE`).
	TillerImage   string `json:"tillerImage,omitempty"`
	// The image version to use with tiller (default: `v2.9.1` or `$TILLER_VERSION`).
	TillerVersion string `json:"tillerVersion,omitempty"`
}

// Settings for operating Helm Operator alongside Flux.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type HelmOperator struct {
	// Whether or not to deploy a helm-operator instance in the same namespace (default: false).
	Enabled               bool `json:"enabled,omitempty"`
	// The image to use with helm-operator (default: `quay.io/weaveworks/helm-operator` or `$HELM_OPERATOR_IMAGE`).
	HelmOperatorImage   string `json:"helmOperatorImage,omitempty"`
	// The image version to use with helm-operator (default: `master-1dfdc61` or `$HELM_OPERATOR_VERSION`).
	HelmOperatorVersion string `json:"helmOperatorVersion,omitempty"`
	// The chart path to use with Helm Operator (default: `.`).
	ChartPath           string `json:"chartPath,omitempty"`
	// The frequency with which to sync Git and the charts (default: the flux `GitPollInterval` or, if not set, `3m0s`).
	GitPollInterval     string `json:"gitPollInterval,omitempty"`
	// The URL of the git repository to use if it is different than the primary flux `GitUrl`.
	GitUrl              string `json:"gitUrl,omitempty"`
}

type FluxStatus struct {
	// Fill me
}
