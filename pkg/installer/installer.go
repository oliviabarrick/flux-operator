package installer

import (
	"fmt"
	crdutils "github.com/ant31/crd-validation/pkg"
	v1alpha1 "github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"os"
	"reflect"
	"strconv"
)

// Represents the configuration for a flux-operator instance.
type FluxOperatorConfig struct {
	// The name to use for flux-operator resources
	Name string
	// The namespace to deploy flux-operator to.
	Namespace string
	// Whether or not flux-operator should be namespace or cluster scoped.
	Cluster bool
	// The service account to assign to flux-operator, otherwise one is created.
	ServiceAccount string
	// A cluster role to assign to flux-operator (default: a role is created with
	// full privileges)
	ClusterRole string
	// Do not change any RBAC settings when creating flux-operator.
	DisableRBAC bool
	// Do not allow flux-operator to create any cluster roles.
	DisableClusterRoles bool
	// Do not allow flux-operator to create any roles.
	DisableRoles bool
	// The default git secret to use for fluxes.
	GitSecret string
	// The flux operator image name
	FluxOperatorImage string
	// The flux operator image version
	FluxOperatorVersion string
	// The flux image name
	FluxImage string
	// The flux image version
	FluxVersion string
	// The helm-operator image name
	HelmOperatorImage string
	// The helm-operator image version
	HelmOperatorVersion string
	// The memcached image name
	MemcachedImage string
	// The memcached image version
	MemcachedVersion string
	// The tiller image name
	TillerImage string
	// The flux image version
	TillerVersion string
	// If set, restricts flux-operator to look for new Fluxes only in the specified
	// namespace.
	FluxNamespace string
}

// Return the name that should be used for flux-operator resources.
func GetName(config FluxOperatorConfig) string {
	if config.Name != "" {
		return config.Name
	} else {
		return "flux-operator"
	}
}

// Return the cluster role name to use, an empty string if none should be used.
func GetClusterRole(config FluxOperatorConfig) string {
	if config.ClusterRole == "" {
		return GetName(config)
	} else {
		return config.ClusterRole
	}
}

// Return the service account name to use, empty string if none should be used.
func GetServiceAccountName(config FluxOperatorConfig) string {
	if config.DisableRBAC == true {
		return ""
	} else if config.ServiceAccount != "" {
		return config.ServiceAccount
	} else {
		return GetName(config)
	}
}

// Return the flux-operator namespace.
func GetNamespace(config FluxOperatorConfig) string {
	if config.Namespace != "" {
		return config.Namespace
	} else {
		return "default"
	}
}

// Return the flux operator image name.
func GetFluxOperatorImage(config FluxOperatorConfig) string {
	image := utils.FluxOperatorImage
	version := "latest"

	if config.FluxOperatorImage != "" {
		image = config.FluxOperatorImage
	}

	if config.FluxOperatorVersion != "" {
		version = config.FluxOperatorVersion
	}

	return fmt.Sprintf("%s:%s", image, version)
}

// Create Flux CRD
func NewFluxCRD(config FluxOperatorConfig) *extensions.CustomResourceDefinition {
	scope := "Namespaced"
	if config.Cluster {
		scope = "Cluster"
	}

	spec := "github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1.Flux"
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		SpecDefinitionName:    spec,
		EnableValidation:      true,
		ResourceScope:         scope,
		Group:                 "flux.codesink.net",
		Kind:                  "Flux",
		Version:               "v1alpha1",
		Plural:                "fluxes",
		GetOpenAPIDefinitions: v1alpha1.GetOpenAPIDefinitions,
	})
}

// Create a FluxHelmRelease CRD
func NewFluxHelmReleaseCRD(FluxOperatorConfig) *extensions.CustomResourceDefinition {
	return &extensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluxhelmreleases.helm.integrations.flux.weave.works",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		Spec: extensions.CustomResourceDefinitionSpec{
			Group: "helm.integrations.flux.weave.works",
			Names: extensions.CustomResourceDefinitionNames{
				Kind:     "FluxHelmRelease",
				ListKind: "FluxHelmReleaseList",
				Plural:   "fluxhelmreleases",
			},
			Scope:   "Namespaced",
			Version: "v1alpha2",
		},
	}
}

// Create a flux-operator deployment
func NewFluxOperatorDeployment(config FluxOperatorConfig) *v1beta1.Deployment {
	replicas := int32(1)

	labels := map[string]string{
		"app": "flux-operator",
	}

	return &v1beta1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux-operator",
			Namespace: GetNamespace(config),
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: GetServiceAccountName(config),
					Containers: []corev1.Container{
						{
							Name:            "flux-operator",
							Image:           GetFluxOperatorImage(config),
							ImagePullPolicy: "IfNotPresent",
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "WATCH_NAMESPACE",
									Value: GetNamespace(config),
								},
								corev1.EnvVar{
									Name:  "GIT_SECRET_NAME",
									Value: config.GitSecret,
								},
								corev1.EnvVar{
									Name:  "FLUX_IMAGE",
									Value: config.FluxImage,
								},
								corev1.EnvVar{
									Name:  "FLUX_VERSION",
									Value: config.FluxVersion,
								},
								corev1.EnvVar{
									Name:  "HELM_OPERATOR_IMAGE",
									Value: config.HelmOperatorImage,
								},
								corev1.EnvVar{
									Name:  "HELM_OPERATOR_VERSION",
									Value: config.HelmOperatorVersion,
								},
								corev1.EnvVar{
									Name:  "MEMCACHED_IMAGE",
									Value: config.MemcachedImage,
								},
								corev1.EnvVar{
									Name:  "MEMCACHED_VERSION",
									Value: config.MemcachedVersion,
								},
								corev1.EnvVar{
									Name:  "TILLER_IMAGE",
									Value: config.TillerImage,
								},
								corev1.EnvVar{
									Name:  "TILLER_VERSION",
									Value: config.TillerVersion,
								},
								corev1.EnvVar{
									Name:  "FLUX_NAMESPACE",
									Value: config.FluxNamespace,
								},
								corev1.EnvVar{
									Name:  "DISABLE_ROLES",
									Value: strconv.FormatBool(config.DisableRoles),
								},
								corev1.EnvVar{
									Name:  "DISABLE_CLUSTER_ROLES",
									Value: strconv.FormatBool(config.DisableClusterRoles),
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU:    resource.MustParse("250m"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create the service account
func NewServiceAccount(config FluxOperatorConfig) *corev1.ServiceAccount {
	if config.DisableRBAC || config.ServiceAccount != "" {
		return nil
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetServiceAccountName(config),
			Namespace: GetNamespace(config),
		},
	}
}

// Create a cluster role
func NewClusterRole(config FluxOperatorConfig) *rbacv1.ClusterRole {
	if config.DisableRBAC || config.ClusterRole != "" {
		return nil
	}

	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: GetClusterRole(config),
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			},
		},
	}
}

// Create the cluster role binding
func NewClusterRoleBinding(config FluxOperatorConfig) *rbacv1.ClusterRoleBinding {
	if config.DisableRBAC {
		return nil
	}

	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: GetName(config),
		},
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind:      "ServiceAccount",
				Name:      GetServiceAccountName(config),
				Namespace: GetNamespace(config),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     GetClusterRole(config),
		},
	}
}

// Create a flux-operator and all dependent resources.
func NewFluxOperator(config FluxOperatorConfig) []runtime.Object {
	return []runtime.Object{
		NewFluxCRD(config), NewFluxHelmReleaseCRD(config), NewServiceAccount(config),
		NewClusterRole(config), NewClusterRoleBinding(config),
		NewFluxOperatorDeployment(config),
	}
}

// Just print the YAML, don't create it in the API.
func DryRun(config FluxOperatorConfig) {
	encoder := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	manifests := NewFluxOperator(config)
	for index, manifest := range manifests {
		if reflect.ValueOf(manifest).IsNil() {
			continue
		}

		err := encoder.Encode(manifest, os.Stdout)
		if err != nil {
			panic(err)
		}

		if index+1 != len(manifests) {
			fmt.Println("---")
		}
	}
}
