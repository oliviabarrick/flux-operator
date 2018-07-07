package stub

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Find all resources that currently exist for the CR.
func ExistingFluxObjects(cr *v1alpha1.Flux) (existing []runtime.Object, err error) {
	lists := []runtime.Object{
		&extensions.DeploymentList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "extensions/v1beta1",
			},
		},
		&corev1.ServiceList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
		},
		&corev1.SecretList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
		},
		&corev1.ServiceAccountList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: "v1",
			},
		},
		&rbacv1.ClusterRoleList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRole",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
		},
		&rbacv1.ClusterRoleBindingList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
		},
		&rbacv1.RoleList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
		},
		&rbacv1.RoleBindingList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
		},
	}

	for _, list := range lists {
		err = ListForFlux(cr, list)
		if err != nil {
			return
		}

		items, _ := meta.ExtractList(list)
		existing = append(existing, items...)
	}

	return
}

// List all resources of a certain type for a CR.
func ListForFlux(cr *v1alpha1.Flux, list sdk.Object) error {
	opts := sdk.WithListOptions(utils.ListOptionsForFlux(cr))

	err := sdk.List(utils.FluxNamespace(cr), list, opts)
	if err != nil {
		return err
	}

	return nil
}
