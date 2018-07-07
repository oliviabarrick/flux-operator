package test_utils

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewFlux() *v1alpha1.Flux {
	return &v1alpha1.Flux{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: v1alpha1.FluxSpec{
			GitUrl:          "git@github.com:justinbarrick/manifests",
			GitBranch:       "master",
			GitPath:         "manifests",
			GitPollInterval: "0m30s",
		},
	}
}
