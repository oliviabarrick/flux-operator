package flux

import (
	"fmt"
	"sort"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
)

func TestMakeFluxArgs(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Args = map[string]string{
		"connect": "ws://fluxcloud/",
	}

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-example",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--connect=ws://fluxcloud/",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsNoArgs(t *testing.T) {
	cr := test_utils.NewFlux()
	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-example",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsArgsOverride(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Args = map[string]string{
		"git-url": "git@github.com:justinbarrick/flux-operator",
	}

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/flux-operator",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-example",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestNewFluxDeployment(t *testing.T) {
	cr := test_utils.NewFlux()
	dep := NewFluxDeployment(cr)
	pod := dep.Spec.Template.Spec

	assert.Equal(t, dep.ObjectMeta.Name, "flux-example")
	assert.Equal(t, dep.ObjectMeta.Namespace, "default")
	assert.Equal(t, pod.ServiceAccountName, "flux-example")
	assert.Equal(t, pod.Volumes[0].VolumeSource.Secret.SecretName, "flux-git-example-deploy")
	assert.Equal(t, *pod.Volumes[0].VolumeSource.Secret.DefaultMode, int32(0400))

	c := pod.Containers[0]
	assert.Equal(t, c.Image, fmt.Sprintf("%s:%s", utils.FluxImage, utils.FluxVersion))

	expectedArgs := MakeFluxArgs(cr)
	sort.Strings(expectedArgs)

	assert.Equal(t, c.Args, expectedArgs)
}

func TestNewFluxDeploymentOverrides(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxImage = "myimage"
	cr.Spec.FluxVersion = "myversion"
	cr.Spec.GitSecret = "mysecret"

	dep := NewFluxDeployment(cr)
	pod := dep.Spec.Template.Spec
	assert.Equal(t, pod.Volumes[0].VolumeSource.Secret.SecretName, "mysecret")
	assert.Equal(t, pod.Containers[0].Image, "myimage:myversion")
}
