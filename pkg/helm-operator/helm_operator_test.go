package helm_operator

import (
	"sort"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
)

func TestMakeHelmOperatorArgs(t *testing.T) {
	cr := test_utils.NewFlux()

	args := MakeHelmOperatorArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-charts-path=./",
		"--charts-sync-interval=0m30s",
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
	cr.Spec.HelmOperator.GitUrl = "example.git"

	args := MakeHelmOperatorArgs(cr)

	expectedArgs := []string{
		"--git-url=example.git",
		"--git-branch=master",
		"--git-charts-path=charts/",
		"--charts-sync-interval=1m0s",
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
	assert.Equal(t, c.Image, "quay.io/weaveworks/helm-operator:master-1dfdc61")

	expectedArgs := MakeHelmOperatorArgs(cr)
	sort.Strings(expectedArgs)

	assert.Equal(t, c.Args, expectedArgs)
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
