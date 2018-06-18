package flux

import (
	"sort"
	"testing"
	"github.com/stretchr/testify/assert"
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
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(args)
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
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(args)
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
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(args)
	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestNewFluxPod(t *testing.T) {
	cr := test_utils.NewFlux()
	pod := NewFluxPod(cr)

	assert.Equal(t, pod.ObjectMeta.Name, "flux-example")
	assert.Equal(t, pod.ObjectMeta.Namespace, "default")
	assert.Equal(t, pod.Spec.ServiceAccountName, "flux-example")
	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.Secret.SecretName, "flux-git-example-deploy")

	c := pod.Spec.Containers[0]
	assert.Equal(t, c.Image, "quay.io/weaveworks/flux:1.4.0")

	expectedArgs := MakeFluxArgs(cr)
	sort.Strings(expectedArgs)
	sort.Strings(c.Args)

	assert.Equal(t, c.Args, expectedArgs)
}

func TestNewFluxPodOverrides(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxImage = "myimage"
	cr.Spec.FluxVersion = "myversion"
	cr.Spec.GitSecret = "mysecret"

	pod := NewFluxPod(cr)

	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.Secret.SecretName, "mysecret")
	assert.Equal(t, pod.Spec.Containers[0].Image, "myimage:myversion")
}
