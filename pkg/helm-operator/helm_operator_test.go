package helm_operator

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sort"
	"testing"
)

func TestMakeHelmOperatorArgs(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.GitPollInterval = ""

	args := MakeHelmOperatorArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-charts-path=./",
		"--chart-sync-interval=3m0s",
		"--git-poll-interval=3m0s",
		"--tiller-namespace=default",
		"--tiller-ip=flux-example-tiller-deploy",
		"--tiller-port=44134",
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeHelmOperatorArgsOverrides(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.HelmOperator.ChartPath = "charts/"
	cr.Spec.HelmOperator.GitPollInterval = "1m0s"
	cr.Spec.HelmOperator.ChartSyncInterval = "1m30s"
	cr.Spec.HelmOperator.GitUrl = "example.git"

	args := MakeHelmOperatorArgs(cr)

	expectedArgs := []string{
		"--git-url=example.git",
		"--git-branch=master",
		"--git-charts-path=charts/",
		"--git-poll-interval=1m0s",
		"--chart-sync-interval=1m30s",
		"--tiller-namespace=default",
		"--tiller-ip=flux-example-tiller-deploy",
		"--tiller-port=44134",
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeHelmOperatorArgsOverridesFromBase(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.GitPollInterval = "0m30s"
	cr.Spec.SyncInterval = "1m30s"

	args := MakeHelmOperatorArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-charts-path=./",
		"--git-poll-interval=0m30s",
		"--chart-sync-interval=1m30s",
		"--tiller-namespace=default",
		"--tiller-ip=flux-example-tiller-deploy",
		"--tiller-port=44134",
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestNewHelmOperatorDeployment(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.HelmOperator.Enabled = true
	dep := NewHelmOperatorDeployment(cr)
	pod := dep.Spec.Template.Spec

	assert.Equal(t, dep.ObjectMeta.Name, "flux-example-helm-operator")
	assert.Equal(t, dep.ObjectMeta.Namespace, "default")
	assert.Equal(t, pod.ServiceAccountName, "flux-example")
	assert.Equal(t, pod.Volumes[0].VolumeSource.Secret.SecretName, "flux-git-example-deploy")
	assert.Equal(t, *pod.Volumes[0].VolumeSource.Secret.DefaultMode, int32(0400))

	c := pod.Containers[0]
	assert.Equal(t, c.Image, fmt.Sprintf("%s:%s", utils.HelmOperatorImage, utils.HelmOperatorVersion))

	expectedArgs := MakeHelmOperatorArgs(cr)
	sort.Strings(expectedArgs)

	assert.Equal(t, c.Args, expectedArgs)
	assert.Equal(t, resource.MustParse("512Mi"), c.Resources.Limits[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("128Mi"), c.Resources.Requests[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("1000m"), c.Resources.Limits[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("250m"), c.Resources.Requests[corev1.ResourceCPU])
}

func TestNewHelmOperatorDeploymentOverrides(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.HelmOperator.Enabled = true
	cr.Spec.HelmOperator.HelmOperatorImage = "myimage"
	cr.Spec.HelmOperator.HelmOperatorVersion = "myversion"
	cr.Spec.GitSecret = "mysecret"

	dep := NewHelmOperatorDeployment(cr)
	pod := dep.Spec.Template.Spec

	assert.Equal(t, pod.Volumes[0].VolumeSource.Secret.SecretName, "mysecret")
	assert.Equal(t, pod.Containers[0].Image, "myimage:myversion")
}

func TestNewHelmOperatorDeploymentDisabledByDefault(t *testing.T) {
	assert.Nil(t, NewHelmOperatorDeployment(test_utils.NewFlux()))
}

func TestNewHelmOperatorDeploymentOverrideResources(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.HelmOperator.Enabled = true
	cr.Spec.HelmOperator.Resources = &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1234Mi"),
			corev1.ResourceCPU:    resource.MustParse("1235m"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1000Mi"),
			corev1.ResourceCPU:    resource.MustParse("1337m"),
		},
	}

	dep := NewHelmOperatorDeployment(cr)
	pod := dep.Spec.Template.Spec
	c := pod.Containers[0]

	assert.Equal(t, resource.MustParse("1234Mi"), c.Resources.Limits[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("1000Mi"), c.Resources.Requests[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("1235m"), c.Resources.Limits[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("1337m"), c.Resources.Requests[corev1.ResourceCPU])
}
