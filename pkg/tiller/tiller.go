package tiller

import (
	"bytes"
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/helm/cmd/helm/installer"
	extensions "k8s.io/api/extensions/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Decode a YAML manifest into `out`.
func TillerManifest(asStr string, out interface{}) (error) {
	err := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(asStr), len(asStr)).Decode(out)
	if err != nil {
		return err
	}

	return nil
}

// Create Tiller installation options from a CR.
func TillerOptions(cr *v1alpha1.Flux) *installer.Options {
	tillerImage := utils.Getenv("TILLER_IMAGE", "gcr.io/kubernetes-helm/tiller")
	if cr.Spec.Tiller.TillerImage != "" {
		tillerImage = cr.Spec.Tiller.TillerImage
	}

	tillerVersion := utils.Getenv("TILLER_VERSION", "v2.9.1")
	if cr.Spec.Tiller.TillerVersion != "" {
		tillerVersion = cr.Spec.Tiller.TillerVersion
	}

	return &installer.Options{
		Namespace: utils.FluxNamespace(cr),
		ServiceAccount: rbac.ServiceAccountName(cr),
		ImageSpec: fmt.Sprintf("%s:%s", tillerImage, tillerVersion),
	}
}

// Create the name for a Tiller instance.
func TillerName(cr *v1alpha1.Flux) string {
	return "tiller-deploy"
}

// Create the ObjectMeta for a Tiller installation manifest.
func NewTillerObjectMeta(cr *v1alpha1.Flux) metav1.ObjectMeta {
	meta := utils.NewObjectMeta(cr, TillerName(cr))
	meta.Labels = map[string]string{
		"app": "helm",
		"name": "tiller",
	}
	return meta
}

// Create a Tiller Deployment manifest.
func NewTillerDeployment(cr *v1alpha1.Flux) (*extensions.Deployment, error) {
	deployment := &extensions.Deployment{}

	asStr, err := installer.DeploymentManifest(TillerOptions(cr))
	if err != nil {
		return nil, err
	}

	err = TillerManifest(asStr, deployment)
	if err != nil {
		return nil, err
	}

	deployment.TypeMeta = metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "extensions/v1beta1",
	}
	deployment.ObjectMeta = NewTillerObjectMeta(cr)
	deployment.Spec.Template.ObjectMeta.Labels = deployment.ObjectMeta.Labels
	return deployment, nil
}

// Create a Tiller Service manifest.
func NewTillerService(cr *v1alpha1.Flux) (*corev1.Service, error) {
	service := &corev1.Service{}

	asStr, err := installer.ServiceManifest(utils.FluxNamespace(cr))
	if err != nil {
		return nil, err
	}

	err = TillerManifest(asStr, service)
	if err != nil {
		return nil, err
	}

	service.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
	service.ObjectMeta = NewTillerObjectMeta(cr)
	service.Spec.Selector = service.ObjectMeta.Labels
	return service, nil
}

// Return all objects required to make tiller.
func NewTiller(cr *v1alpha1.Flux) (objects []runtime.Object, err error) {
	if ! cr.Spec.Tiller.Enabled {
		return
	}

	deployment, err := NewTillerDeployment(cr)
	if err != nil {
		return
	}

	service, err := NewTillerService(cr)
	if err != nil {
		return
	}

	objects = append(objects, deployment, service)
	return
}
