package rbac

import (
	"testing"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
)

func TestNewServiceAccount(t *testing.T) {
	sa := NewServiceAccount(test_utils.NewFlux())

	assert.Equal(t, sa.ObjectMeta.Name, "flux-example")
	assert.Equal(t, sa.ObjectMeta.Namespace, "default")
}

func TestNewRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Role.Enabled = true

	role := NewRole(cr)

	defaultRules := []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs: []string{"*"},
		},
	}

	assert.Equal(t, role.Rules, defaultRules)
}

func TestNewCustomRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Role.Enabled = true
	cr.Spec.Role.Rules = []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"myapi"},
			Resources: []string{"pods"},
			Verbs: []string{"GET"},
		},
	}

	role := NewRole(cr)
	assert.Equal(t, role.Rules, cr.Spec.Role.Rules)
}

func TestNewRoleDisabledByDefault(t *testing.T) {
	cr := test_utils.NewFlux()
	role := NewRole(cr)

	assert.Equal(t, cr.Spec.Role.Enabled, false)
	assert.Nil(t, role)
}

func TestNewRoleBinding(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Role.Enabled = true

	roleBinding := NewRoleBinding(cr)

	assert.Equal(t, roleBinding.Subjects[0].Kind, "ServiceAccount")
	assert.Equal(t, roleBinding.Subjects[0].Name, "flux-example")
	assert.Equal(t, roleBinding.Subjects[0].Namespace, "default")

	assert.Equal(t, roleBinding.RoleRef.APIGroup, "rbac.authorization.k8s.io")
	assert.Equal(t, roleBinding.RoleRef.Kind, "Role")
	assert.Equal(t, roleBinding.RoleRef.Name, "flux-example")
}

func TestNewRoleBindingDefaultNil(t *testing.T) {
	assert.Nil(t, NewRoleBinding(test_utils.NewFlux()))
}

func TestNewClusterRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.ClusterRole.Enabled = true

	clusterRole := NewClusterRole(cr)

	defaultRules := []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs: []string{"*"},
		},
		rbacv1.PolicyRule{
			NonResourceURLs: []string{"*"},
			Verbs: []string{"*"},
		},
	}

	assert.Equal(t, clusterRole.Rules, defaultRules)
}

func TestNewCustomClusterRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.ClusterRole.Enabled = true
	cr.Spec.ClusterRole.Rules = []rbacv1.PolicyRule{
		rbacv1.PolicyRule{
			APIGroups: []string{"*"},
			Resources: []string{"*"},
			Verbs: []string{"*"},
		},
		rbacv1.PolicyRule{
			NonResourceURLs: []string{"*"},
			Verbs: []string{"*"},
		},
	}

	clusterRole := NewClusterRole(cr)
	assert.Equal(t, clusterRole.Rules, cr.Spec.ClusterRole.Rules)
}

func TestNewClusterRoleDisabledByDefault(t *testing.T) {
	cr := test_utils.NewFlux()
	clusterRole := NewClusterRole(cr)

	assert.Equal(t, cr.Spec.ClusterRole.Enabled, false)
	assert.Nil(t, clusterRole)
}

func TestNewClusterRoleBinding(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.ClusterRole.Enabled = true

	roleBinding := NewClusterRoleBinding(cr)

	assert.Equal(t, roleBinding.Subjects[0].Kind, "ServiceAccount")
	assert.Equal(t, roleBinding.Subjects[0].Name, "flux-example")
	assert.Equal(t, roleBinding.Subjects[0].Namespace, "default")

	assert.Equal(t, roleBinding.RoleRef.APIGroup, "rbac.authorization.k8s.io")
	assert.Equal(t, roleBinding.RoleRef.Kind, "ClusterRole")
	assert.Equal(t, roleBinding.RoleRef.Name, "flux-example")
}

func TestNewClusterRoleBindingDefaultNil(t *testing.T) {
	assert.Nil(t, NewClusterRoleBinding(test_utils.NewFlux()))
}

func TestFluxRolesDefault(t *testing.T) {
	cr := test_utils.NewFlux()
	objects := FluxRoles(cr)
	assert.Equal(t, len(objects), 1)
	_ = objects[0].(*corev1.ServiceAccount)
}

func TestFluxRolesWithRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Role.Enabled = true
	objects := FluxRoles(cr)
	assert.Equal(t, len(objects), 3)
	_ = objects[0].(*corev1.ServiceAccount)
	_ = objects[1].(*rbacv1.Role)
	_ = objects[2].(*rbacv1.RoleBinding)
}

func TestFluxRolesWithClusterRole(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.ClusterRole.Enabled = true
	objects := FluxRoles(cr)
	assert.Equal(t, len(objects), 3)
	_ = objects[0].(*corev1.ServiceAccount)
	_ = objects[1].(*rbacv1.ClusterRole)
	_ = objects[2].(*rbacv1.ClusterRoleBinding)
}

func TestFluxRolesWithBoth(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Role.Enabled = true
	cr.Spec.ClusterRole.Enabled = true
	objects := FluxRoles(cr)
	assert.Equal(t, len(objects), 5)
	_ = objects[0].(*corev1.ServiceAccount)
	_ = objects[1].(*rbacv1.Role)
	_ = objects[2].(*rbacv1.RoleBinding)
	_ = objects[3].(*rbacv1.ClusterRole)
	_ = objects[4].(*rbacv1.ClusterRoleBinding)
}
