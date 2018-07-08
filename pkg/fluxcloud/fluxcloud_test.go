package fluxcloud

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"testing"
)

func TestFluxcloudName(t *testing.T) {
	cr := test_utils.NewFlux()
	name := FluxcloudName(cr)
	assert.Equal(t, name, "flux-"+cr.ObjectMeta.Name+"-fluxcloud")
}

func TestFluxcloudImage(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, FluxcloudImage(cr), utils.FluxcloudImage+":"+utils.FluxcloudVersion)

	cr.Spec.FluxCloud.FluxCloudImage = "myimage"
	assert.Equal(t, FluxcloudImage(cr), "myimage:"+utils.FluxcloudVersion)

	cr.Spec.FluxCloud.FluxCloudVersion = "myversion"
	assert.Equal(t, FluxcloudImage(cr), "myimage:myversion")

	os.Setenv("FLUXCLOUD_IMAGE", "fluxcloud")
	defer os.Setenv("FLUXCLOUD_IMAGE", "")
	assert.Equal(t, FluxcloudImage(cr), "fluxcloud:myversion")

	os.Setenv("FLUXCLOUD_VERSION", "fluxcloudversion")
	defer os.Setenv("FLUXCLOUD_VERSION", "")
	assert.Equal(t, FluxcloudImage(cr), "fluxcloud:fluxcloudversion")
}

func getEnvVar(name string, vars []corev1.EnvVar) string {
	for _, envVar := range vars {
		if envVar.Name == name {
			return envVar.Value
		}
	}

	return ""
}

func TestNewFluxcloudDeployment(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxCloud.Enabled = true
	cr.Spec.FluxCloud.SlackURL = "https://slack/"
	cr.Spec.FluxCloud.SlackChannel = "#channel"
	cr.Spec.FluxCloud.SlackUsername = "My User"
	cr.Spec.FluxCloud.SlackIconEmoji = ":hello:"
	cr.Spec.FluxCloud.GithubURL = "https://github.com/"

	dep := NewFluxcloudDeployment(cr)
	pod := dep.Spec.Template.Spec

	assert.Equal(t, dep.ObjectMeta.Name, FluxcloudName(cr))
	assert.Equal(t, dep.ObjectMeta.Namespace, "default")

	c := pod.Containers[0]
	assert.Equal(t, c.Image, fmt.Sprintf("%s:%s", utils.FluxcloudImage, utils.FluxcloudVersion))
	assert.Equal(t, getEnvVar("SLACK_URL", c.Env), cr.Spec.FluxCloud.SlackURL)
	assert.Equal(t, getEnvVar("SLACK_CHANNEL", c.Env), cr.Spec.FluxCloud.SlackChannel)
	assert.Equal(t, getEnvVar("SLACK_USERNAME", c.Env), cr.Spec.FluxCloud.SlackUsername)
	assert.Equal(t, getEnvVar("SLACK_ICON_EMOJI", c.Env), cr.Spec.FluxCloud.SlackIconEmoji)
	assert.Equal(t, getEnvVar("GITHUB_URL", c.Env), cr.Spec.FluxCloud.GithubURL)
}

func TestNewFluxcloudDeploymentDisabled(t *testing.T) {
	cr := test_utils.NewFlux()

	dep := NewFluxcloudDeployment(cr)
	assert.Nil(t, dep)
}

func TestNewFluxcloudService(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxCloud.Enabled = true

	service := NewFluxcloudService(cr)

	assert.Equal(t, service.ObjectMeta.Name, FluxcloudName(cr))
	assert.Equal(t, service.ObjectMeta.Namespace, "default")

	assert.Equal(t, service.Spec.Ports[0].Name, "fluxcloud")
	assert.Equal(t, service.Spec.Ports[0].Port, int32(80))
	assert.Equal(t, service.Spec.Ports[0].TargetPort, intstr.FromInt(3031))
	assert.Equal(t, service.Spec.Selector["name"], FluxcloudName(cr))
}

func TestNewFluxcloudServiceDisabled(t *testing.T) {
	cr := test_utils.NewFlux()

	service := NewFluxcloudService(cr)
	assert.Nil(t, service)
}

func TestNewFluxcloud(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxCloud.Enabled = true

	objects := NewFluxcloud(cr)

	assert.Equal(t, len(objects), 2)
	_ = objects[0].(*extensions.Deployment)
	_ = objects[1].(*corev1.Service)
}

func TestNewFluxcloudDisabled(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxCloud.Enabled = false

	objects := NewFluxcloud(cr)

	assert.Equal(t, len(objects), 2)
	assert.Nil(t, objects[0])
	assert.Nil(t, objects[1])
}
