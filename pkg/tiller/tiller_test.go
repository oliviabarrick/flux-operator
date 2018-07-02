package tiller

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"k8s.io/helm/cmd/helm/installer"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
)

func TestTillerName(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, TillerName(cr), "flux-example-tiller-deploy")
}

func TestTillerOptions(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, *TillerOptions(cr), installer.Options{
		Namespace: utils.FluxNamespace(cr),
		ServiceAccount: rbac.ServiceAccountName(cr),
		ImageSpec: "gcr.io/kubernetes-helm/tiller:v2.9.1",
	})
}

func TestTillerImageOverride(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Tiller.TillerImage = "tiller"
	cr.Spec.Tiller.TillerVersion = "version"

	assert.Equal(t, *TillerOptions(cr), installer.Options{
		Namespace: utils.FluxNamespace(cr),
		ServiceAccount: rbac.ServiceAccountName(cr),
		ImageSpec: "tiller:version",
	})
}

func TestNewTillerObjectMeta(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, NewTillerObjectMeta(cr).Name, TillerName(cr))
	assert.Equal(t, NewTillerObjectMeta(cr).Namespace, utils.FluxNamespace(cr))
}

func TestNewTillerDeployment(t *testing.T) {
	cr := test_utils.NewFlux()

	deployment, err := NewTillerDeployment(cr)
	assert.Nil(t, err)
	assert.Equal(t, deployment.ObjectMeta.Name, TillerName(cr))
	c := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, deployment.ObjectMeta.Namespace, utils.FluxNamespace(cr))
	assert.Equal(t, c.Name, "tiller")
	assert.Equal(t, c.Image, TillerOptions(cr).ImageSpec)
	assert.Equal(t, string(c.ImagePullPolicy), "IfNotPresent")
	assert.Equal(t, c.Env[0].Name, "TILLER_NAMESPACE")
	assert.Equal(t, c.Env[0].Value, utils.FluxNamespace(cr))
	assert.Equal(t, c.Ports[0].ContainerPort, int32(44134))
	assert.Equal(t, c.Ports[0].Name, "tiller")
	assert.Equal(t, c.Ports[1].ContainerPort, int32(44135))
	assert.Equal(t, deployment.Spec.Template.ObjectMeta.Labels["app"], "helm")
	assert.Equal(t, deployment.Spec.Template.ObjectMeta.Labels["name"], "tiller")
}

func TestNewTillerService(t *testing.T) {
	cr := test_utils.NewFlux()

	service, err := NewTillerService(cr)
	assert.Nil(t, err)
	assert.Equal(t, service.ObjectMeta.Name, TillerName(cr))
	assert.Equal(t, service.ObjectMeta.Namespace, utils.FluxNamespace(cr))
	assert.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	assert.Equal(t, service.Spec.Ports[0].Port, int32(44134))
	assert.Equal(t, service.Spec.Ports[0].Name, "tiller")
	assert.Equal(t, service.Spec.Selector["app"], "helm")
	assert.Equal(t, service.Spec.Selector["name"], "tiller")
}

func TestNewTillerDefaultDisabled(t *testing.T) {
	cr := test_utils.NewFlux()

	objects, err := NewTiller(cr)
	assert.Nil(t, err)
	assert.Equal(t, len(objects), 0)
}

func TestNewTillerEnabled(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Tiller.Enabled = true

	objects, err := NewTiller(cr)
	assert.Nil(t, err)
	assert.Equal(t, len(objects), 2)
	_ = objects[0].(*extensions.Deployment)
	_ = objects[1].(*corev1.Service)
}
