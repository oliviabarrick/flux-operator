package utils

import (
	"fmt"
	"os"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Getenv(name, value string) string {
	ret := os.Getenv(name)
	if ret == "" {
		return value
	}
	return ret
}

func NewObjectMeta(cr *v1alpha1.Flux, name string) metav1.ObjectMeta {
	if name == "" {
		name = fmt.Sprintf("flux-%s", cr.Name)
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: cr.Spec.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(cr, schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Kind:    "Flux",
			}),
		},
	}
}


