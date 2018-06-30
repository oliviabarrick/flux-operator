package rbac

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/justinbarrick/flux-operator/pkg/utils"
)

func ServiceAccountName(cr *v1alpha1.Flux) string {
	return fmt.Sprintf("flux-%s", cr.Name)
}

func NewServiceAccount(cr *v1alpha1.Flux) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, ServiceAccountName(cr)),
	}
}

func NewClusterRole(cr *v1alpha1.Flux) *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, fmt.Sprintf("flux-%s", cr.Name)),
	}

	if cr.Spec.ClusterRole.Enabled == false || utils.BoolEnv("DISABLE_CLUSTER_ROLES") {
		clusterRole.Rules = []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs: []string{"get", "watch", "list"},
			},
		}
	} else if len(cr.Spec.ClusterRole.Rules) > 0 {
		clusterRole.Rules = cr.Spec.ClusterRole.Rules
	} else {
		clusterRole.Rules = []rbacv1.PolicyRule{
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
	}

	return clusterRole
}

func NewClusterRoleBinding(cr *v1alpha1.Flux) *rbacv1.ClusterRoleBinding {
	serviceAccount := fmt.Sprintf("flux-%s", cr.Name)
	meta := utils.NewObjectMeta(cr, fmt.Sprintf("flux-%s", cr.Name))

	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: meta,
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: serviceAccount,
				Namespace: utils.FluxNamespace(cr),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind: "ClusterRole",
			Name: meta.Name,
		},
	}
}

func NewRole(cr *v1alpha1.Flux) *rbacv1.Role {
	if cr.Spec.Role.Enabled == false || utils.BoolEnv("DISABLE_ROLES") {
		return nil
	}

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, ""),
	}

	if len(cr.Spec.Role.Rules) > 0 {
		role.Rules = cr.Spec.Role.Rules
	} else {
		role.Rules = []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs: []string{"*"},
			},
		}
	}

	return role
}

func NewRoleBinding(cr *v1alpha1.Flux) *rbacv1.RoleBinding {
	if cr.Spec.Role.Enabled == false || utils.BoolEnv("DISABLE_ROLES") {
		return nil
	}

	meta := utils.NewObjectMeta(cr, "")

	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: meta,
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: ServiceAccountName(cr),
				Namespace: utils.FluxNamespace(cr),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind: "Role",
			Name: meta.Name,
		},
	}
}

func FluxRoles(cr *v1alpha1.Flux) (objects []runtime.Object) {
	objects = append(objects, NewServiceAccount(cr))

	role := NewRole(cr)
	if role != nil {
		objects = append(objects, role)
	}

	roleBinding := NewRoleBinding(cr)
	if roleBinding != nil {
		objects = append(objects, roleBinding)
	}

	clusterRole := NewClusterRole(cr)
	if clusterRole != nil {
		objects = append(objects, clusterRole)
	}

	clusterRoleBinding := NewClusterRoleBinding(cr)
	if clusterRoleBinding != nil {
		objects = append(objects, clusterRoleBinding)
	}

	return objects
}
