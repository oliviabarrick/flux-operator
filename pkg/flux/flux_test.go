package flux

import (
	"sort"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newFlux() *v1alpha1.Flux {
	return &v1alpha1.Flux{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example",
			Namespace: "default",
		},
		Spec: v1alpha1.FluxSpec{
			GitUrl: "git@github.com:justinbarrick/manifests",
			GitBranch: "master",
			GitPath: "manifests",
			GitPollInterval: "0m30s",
		},
	}
}

func TestMakeFluxArgs(t *testing.T) {
	cr := newFlux()
	cr.Spec.Args = map[string]string{
		"connect": "ws://fluxcloud/",
	}

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-master",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--connect=ws://fluxcloud/",
	}

	sort.Strings(args)
	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsNoArgs(t *testing.T) {
	args := MakeFluxArgs(newFlux())

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-master",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
	}

	sort.Strings(args)
	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsArgsOverride(t *testing.T) {
	cr := newFlux()
	cr.Spec.Args = map[string]string{
		"git-url": "git@github.com:justinbarrick/flux-operator",
	}

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/flux-operator",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-master",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
	}

	sort.Strings(args)
	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestNewFluxPod(t *testing.T) {
	cr := newFlux()
	pod := NewFluxPod(cr)

	assert.Equal(t, pod.ObjectMeta.Name, "flux-example")
	assert.Equal(t, pod.ObjectMeta.Namespace, "default")
	assert.Equal(t, pod.Spec.ServiceAccountName, "flux")
	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.Secret.SecretName, "flux-git-deploy")

	c := pod.Spec.Containers[0]
	assert.Equal(t, c.Image, "quay.io/weaveworks/flux:1.2.3")

	expectedArgs := MakeFluxArgs(cr)
	sort.Strings(expectedArgs)
	sort.Strings(c.Args)

	assert.Equal(t, c.Args, expectedArgs)
}

func TestNewFluxPodOverrides(t *testing.T) {
	cr := newFlux()
	cr.Spec.FluxImage = "myimage"
	cr.Spec.FluxVersion = "myversion"
	cr.Spec.GitSecret = "mysecret"

	pod := NewFluxPod(cr)

	assert.Equal(t, pod.Spec.Volumes[0].VolumeSource.Secret.SecretName, "mysecret")
	assert.Equal(t, pod.Spec.Containers[0].Image, "myimage:myversion")
}
