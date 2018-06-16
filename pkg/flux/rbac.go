package flux

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewServiceAccount(cr *v1alpha1.Flux) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: NewObjectMeta(cr, ""),
	}
}

func NewClusterRole(cr *v1alpha1.Flux) *rbacv1.ClusterRole {
	if cr.Spec.ClusterRole.Enabled == false {
		return nil
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: NewObjectMeta(cr, fmt.Sprintf("flux-%s-%s", cr.Namespace, cr.Name)),
	}

	if len(cr.Spec.ClusterRole.Rules) > 0 {
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
	if cr.Spec.ClusterRole.Enabled == false {
		return nil
	}

	serviceAccount := fmt.Sprintf("flux-%s", cr.Name)
	meta := NewObjectMeta(cr, fmt.Sprintf("flux-%s-%s", cr.Namespace, cr.Name))

	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: meta,
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: serviceAccount,
				Namespace: cr.Namespace,
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
	if cr.Spec.Role.Enabled == false {
		return nil
	}

	role := &rbacv1.Role{
		ObjectMeta: NewObjectMeta(cr, ""),
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
			rbacv1.PolicyRule{
				NonResourceURLs: []string{"*"},
				Verbs: []string{"*"},
			},
		}
	}

	return role
}

func NewRoleBinding(cr *v1alpha1.Flux) *rbacv1.RoleBinding {
	if cr.Spec.Role.Enabled == false {
		return nil
	}

	meta := NewObjectMeta(cr, "")

	return &rbacv1.RoleBinding{
		ObjectMeta: meta,
		Subjects: []rbacv1.Subject{
			rbacv1.Subject{
				Kind: "ServiceAccount",
				Name: meta.Name,
				Namespace: cr.Namespace,
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
