package installer

import (
	"fmt"
	"strconv"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

func TestGetName(t *testing.T) {
	assert.Equal(t, "flux-operator", GetName(FluxOperatorConfig{}))
	assert.Equal(t, "my-name", GetName(FluxOperatorConfig{Name: "my-name"}))
}

func TestGetClusterRole(t *testing.T) {
	assert.Equal(t, GetName(FluxOperatorConfig{}), GetClusterRole(FluxOperatorConfig{}))
	assert.Equal(t, "my-role", GetClusterRole(FluxOperatorConfig{ClusterRole: "my-role"}))
}

func TestGetServiceAccountNameName(t *testing.T) {
	assert.Equal(t, GetName(FluxOperatorConfig{}), GetServiceAccountName(FluxOperatorConfig{}))
	assert.Equal(t, "my-name", GetServiceAccountName(FluxOperatorConfig{Name: "my-name"}))

	assert.Equal(t, "", GetServiceAccountName(FluxOperatorConfig{
		DisableRBAC: true,
		Name: "my-name",
	}))
}

func TestGetNamespace(t *testing.T) {
	assert.Equal(t, "default", GetNamespace(FluxOperatorConfig{}))
	assert.Equal(t, "my-namespace", GetNamespace(FluxOperatorConfig{Namespace: "my-namespace"}))
}

func TestGetFluxOperatorImage(t *testing.T) {
	assert.Equal(t, fmt.Sprintf("%s:latest", utils.FluxOperatorImage), GetFluxOperatorImage(FluxOperatorConfig{}))
	assert.Equal(t, "hello:latest", GetFluxOperatorImage(FluxOperatorConfig{
		FluxOperatorImage: "hello",
	}))
	assert.Equal(t, fmt.Sprintf("%s:hello", utils.FluxOperatorImage), GetFluxOperatorImage(FluxOperatorConfig{
		FluxOperatorVersion: "hello",
	}))
}

func TestNewFluxCRD(t *testing.T) {
	fluxCrd := NewFluxCRD(FluxOperatorConfig{})
	assert.Equal(t, extensions.ResourceScope("Namespaced"), fluxCrd.Spec.Scope)
	assert.Equal(t, "Flux", fluxCrd.Spec.Names.Kind)
	assert.Equal(t, "fluxes", fluxCrd.Spec.Names.Plural)
	assert.Equal(t, "v1alpha1", fluxCrd.Spec.Version)
	assert.Equal(t, "flux.codesink.net", fluxCrd.Spec.Group)

	fluxCrd = NewFluxCRD(FluxOperatorConfig{Cluster: true})
	assert.Equal(t, extensions.ResourceScope("Cluster"), fluxCrd.Spec.Scope)
}

func TestNewFluxHelmReleaseCRD(t *testing.T) {
	fluxCrd := NewFluxHelmReleaseCRD(FluxOperatorConfig{})
	assert.Equal(t,  extensions.ResourceScope("Namespaced"), fluxCrd.Spec.Scope)
	assert.Equal(t, "FluxHelmRelease", fluxCrd.Spec.Names.Kind)
	assert.Equal(t, "fluxhelmreleases", fluxCrd.Spec.Names.Plural)
	assert.Equal(t, "FluxHelmReleaseList", fluxCrd.Spec.Names.ListKind)
	assert.Equal(t, "v1alpha2", fluxCrd.Spec.Version)
	assert.Equal(t, "helm.integrations.flux.weave.works", fluxCrd.Spec.Group)
}

func TestNewFluxOperatorDeploymentDefaults(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{})
}

func TestNewFluxOperatorDeploymentGitSecret(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		GitSecret: "my-secret",
	})
}

func TestNewFluxOperatorDeploymentFluxImage(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		FluxImage: "myimage",
		FluxVersion: "myversion",
	})
}

func TestNewFluxOperatorDeploymentHelmOperatorImage(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		HelmOperatorImage: "myimage",
		HelmOperatorVersion: "myversion",
	})
}

func TestNewFluxOperatorDeploymentMemcachedImage(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		MemcachedImage: "myimage",
		MemcachedVersion: "myversion",
	})
}

func TestNewFluxOperatorDeploymentTillerImage(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		TillerImage: "myimage",
		TillerVersion: "myversion",
	})
}

func TestNewFluxOperatorDeploymentFluxNamespace(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		FluxNamespace: "namespace",
	})
}

func TestNewFluxOperatorDeploymentDisableRoles(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		DisableRoles: true,
	})
}

func TestNewFluxOperatorDeploymentDisableClusterRoles(t *testing.T) {
	testFluxOperatorDeployment(t, FluxOperatorConfig{
		DisableClusterRoles: true,
	})
}

func testFluxOperatorDeployment(t *testing.T, config FluxOperatorConfig) {
	fluxOp := NewFluxOperatorDeployment(config)

	labels := map[string]string{
		"app": "flux-operator",
	}

	envVars := fluxOp.Spec.Template.Spec.Containers[0].Env
	assert.Equal(t, int32(1), *fluxOp.Spec.Replicas)
	assert.Equal(t, labels, fluxOp.Spec.Selector.MatchLabels)
	assert.Equal(t, labels, fluxOp.Spec.Template.ObjectMeta.Labels)
	assert.Equal(t, GetServiceAccountName(config), fluxOp.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(t, GetFluxOperatorImage(config), fluxOp.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, GetNamespace(config), getEnvVar("WATCH_NAMESPACE", envVars))
	assert.Equal(t, config.GitSecret, getEnvVar("GIT_SECRET_NAME", envVars))
	assert.Equal(t, config.FluxImage, getEnvVar("FLUX_IMAGE", envVars))
	assert.Equal(t, config.FluxVersion, getEnvVar("FLUX_VERSION", envVars))
	assert.Equal(t, config.HelmOperatorImage, getEnvVar("HELM_OPERATOR_IMAGE", envVars))
	assert.Equal(t, config.HelmOperatorVersion, getEnvVar("HELM_OPERATOR_VERSION", envVars))
	assert.Equal(t, config.MemcachedImage, getEnvVar("MEMCACHED_IMAGE", envVars))
	assert.Equal(t, config.MemcachedVersion, getEnvVar("MEMCACHED_VERSION", envVars))
	assert.Equal(t, config.TillerImage, getEnvVar("TILLER_IMAGE", envVars))
	assert.Equal(t, config.TillerVersion, getEnvVar("TILLER_VERSION", envVars))
	assert.Equal(t, config.FluxNamespace, getEnvVar("FLUX_NAMESPACE", envVars))
	assert.Equal(t, strconv.FormatBool(config.DisableRoles), getEnvVar("DISABLE_ROLES", envVars))
	assert.Equal(t, strconv.FormatBool(config.DisableClusterRoles), getEnvVar("DISABLE_CLUSTER_ROLES", envVars))
}

func TestNewServiceAccount(t *testing.T) {
	config := FluxOperatorConfig{}
	sa := NewServiceAccount(config)
	assert.Equal(t, GetServiceAccountName(config), sa.Name)
	assert.Equal(t, GetNamespace(config), sa.Namespace)
}

func TestNewServiceAccountRBACDisabled(t *testing.T) {
	assert.Nil(t, NewServiceAccount(FluxOperatorConfig{
		DisableRBAC: true,
	}))
}

func TestNewServiceAccountAlreadyExists(t *testing.T) {
	assert.Nil(t, NewServiceAccount(FluxOperatorConfig{
		ServiceAccount: "hello",
	}))
}

func TestClusterRole(t *testing.T) {
	config := FluxOperatorConfig{}
	clusterRole := NewClusterRole(config)
	assert.Equal(t, GetClusterRole(config), clusterRole.ObjectMeta.Name)
	assert.Equal(t, []string{"*"}, clusterRole.Rules[0].APIGroups)
	assert.Equal(t, []string{"*"}, clusterRole.Rules[0].Resources)
	assert.Equal(t, []string{"*"}, clusterRole.Rules[0].Verbs)

	assert.Equal(t, []string{"*"}, clusterRole.Rules[1].NonResourceURLs)
	assert.Equal(t, []string{"*"}, clusterRole.Rules[1].Verbs)
}

func TestClusterRoleWithClusterRoleSet(t *testing.T) {
	assert.Nil(t, NewClusterRole(FluxOperatorConfig{
		ClusterRole: "hello",
	}))
}

func TestClusterRoleWithRBACDisabled(t *testing.T) {
	assert.Nil(t, NewClusterRole(FluxOperatorConfig{
		DisableRBAC: true,
	}))
}

func TestClusterRoleBinding(t *testing.T) {
	testClusterRoleBinding(t, FluxOperatorConfig{})
}

func TestClusterRoleBindingClusterRoleSet(t *testing.T) {
	testClusterRoleBinding(t, FluxOperatorConfig{
		ClusterRole: "hello",
	})
}

func TestClusterRoleBindingServiceAccountSet(t *testing.T) {
	testClusterRoleBinding(t, FluxOperatorConfig{
		ServiceAccount: "hello",
	})
}

func TestClusterRoleBindingDisableRBAC(t *testing.T) {
	assert.Nil(t, NewClusterRoleBinding(FluxOperatorConfig{
		DisableRBAC: true,
	}))
}

func testClusterRoleBinding(t *testing.T, config FluxOperatorConfig) {
	clusterRoleBinding := NewClusterRoleBinding(config)
	assert.Equal(t, GetName(config), clusterRoleBinding.ObjectMeta.Name)
	assert.Equal(t, GetClusterRole(config), clusterRoleBinding.RoleRef.Name)
	assert.Equal(t, "ClusterRole", clusterRoleBinding.RoleRef.Kind)
	assert.Equal(t, "ServiceAccount", clusterRoleBinding.Subjects[0].Kind)
	assert.Equal(t, GetServiceAccountName(config), clusterRoleBinding.Subjects[0].Name)
	assert.Equal(t, GetNamespace(config), clusterRoleBinding.Subjects[0].Namespace)
}

func TestNewFluxOperator(t *testing.T) {
	objs := NewFluxOperator(FluxOperatorConfig{})
	_ = objs[0].(*extensions.CustomResourceDefinition)
	_ = objs[1].(*extensions.CustomResourceDefinition)
	_ = objs[2].(*corev1.ServiceAccount)
	_ = objs[3].(*rbacv1.ClusterRole)
	_ = objs[4].(*rbacv1.ClusterRoleBinding)
	_ = objs[5].(*v1beta1.Deployment)
}

func getEnvVar(name string, vars []corev1.EnvVar) string {
	for _, envvar := range vars {
		if envvar.Name == name {
			return envvar.Value
		}
	}

	return ""
}
