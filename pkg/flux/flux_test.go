package flux

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/fluxcloud"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"sort"
	"testing"
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
		"--sync-interval=5m00s",
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
		"--sync-interval=5m00s",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsOverrideInterval(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.GitPollInterval = "0m30s"
	cr.Spec.SyncInterval = "1m30s"

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-example",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--sync-interval=1m30s",
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
		"--sync-interval=5m00s",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestMakeFluxArgsFluxcloudEnabled(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.FluxCloud.Enabled = true

	args := MakeFluxArgs(cr)

	expectedArgs := []string{
		"--git-url=git@github.com:justinbarrick/manifests",
		"--git-branch=master",
		"--git-sync-tag=flux-sync-example",
		"--git-path=manifests",
		"--git-poll-interval=0m30s",
		"--sync-interval=5m00s",
		"--k8s-secret-name=flux-git-example-deploy",
		"--ssh-keygen-dir=/etc/fluxd/",
		"--memcached-hostname=" + memcached.MemcachedName(cr),
		"--connect=ws://" + fluxcloud.FluxcloudName(cr) + "/",
	}

	sort.Strings(expectedArgs)

	assert.Equal(t, args, expectedArgs)
}

func TestNewFluxDeployment(t *testing.T) {
	cr := test_utils.NewFlux()
	dep := NewFluxDeployment(cr)
	pod := dep.Spec.Template.Spec

	assert.Equal(t, dep.Spec.Selector.MatchLabels["name"], "flux")
	assert.Equal(t, dep.Spec.Template.ObjectMeta.Labels["name"], "flux")
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

	assert.Equal(t, resource.MustParse("512Mi"), c.Resources.Limits[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("128Mi"), c.Resources.Requests[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("500m"), c.Resources.Limits[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("250m"), c.Resources.Requests[corev1.ResourceCPU])
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

func TestNewFluxDeploymentOverrideResources(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Resources = &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1234Mi"),
			corev1.ResourceCPU:    resource.MustParse("1235m"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("1000Mi"),
			corev1.ResourceCPU:    resource.MustParse("1337m"),
		},
	}

	dep := NewFluxDeployment(cr)
	pod := dep.Spec.Template.Spec
	c := pod.Containers[0]

	assert.Equal(t, resource.MustParse("1234Mi"), c.Resources.Limits[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("1000Mi"), c.Resources.Requests[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("1235m"), c.Resources.Limits[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("1337m"), c.Resources.Requests[corev1.ResourceCPU])
}

func TestKnownHostsName(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	assert.Equal(t, fmt.Sprintf("flux-git-%s-known-hosts", cr.Name), KnownHostsName(cr))
}

func TestKnownHostsNameNone(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = ``

	assert.Equal(t, "", KnownHostsName(cr))
}

func TestKnownHostsNameFromEvironment(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = ``

	os.Setenv("KNOWN_HOSTS_CONFIGMAP", "my-configmap")
	defer os.Unsetenv("KNOWN_HOSTS_CONFIGMAP")

	assert.Equal(t, "my-configmap", KnownHostsName(cr))
}

func TestKnownHostsNameOverridesEvironment(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	os.Setenv("KNOWN_HOSTS_CONFIGMAP", "my-configmap")
	defer os.Unsetenv("KNOWN_HOSTS_CONFIGMAP")

	assert.Equal(t, fmt.Sprintf("flux-git-%s-known-hosts", cr.Name), KnownHostsName(cr))
}

func TestNewFluxKnownHosts(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	knownHosts := NewFluxKnownHosts(cr)
	assert.Equal(t, KnownHostsName(cr), knownHosts.ObjectMeta.Name)
	assert.Equal(t, cr.Spec.KnownHosts, knownHosts.Data["known_hosts"])
}

func TestNewFluxKnownHostsEmpty(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = ""

	assert.Nil(t, NewFluxKnownHosts(cr))
}

func TestMakeGitVolumes(t *testing.T) {
	cr := test_utils.NewFlux()
	volumes, volumeMounts := MakeGitVolumes(cr)
	assert.Equal(t, 1, len(volumes))
	assert.Equal(t, volumes[0].Name, "git-key")
	assert.Equal(t, 1, len(volumeMounts))
	assert.Equal(t, volumeMounts[0].Name, "git-key")
}

func TestMakeGitVolumesWithKnownHosts(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	volumes, volumeMounts := MakeGitVolumes(cr)
	assert.Equal(t, 2, len(volumes))
	assert.Equal(t, volumes[0].Name, "git-key")
	assert.Equal(t, volumes[1].Name, "known-hosts")
	assert.Equal(t, volumes[1].VolumeSource.ConfigMap.LocalObjectReference.Name, KnownHostsName(cr))
	assert.Equal(t, 2, len(volumeMounts))
	assert.Equal(t, volumeMounts[0].Name, "git-key")
	assert.Equal(t, volumeMounts[1].Name, "known-hosts")
	assert.Equal(t, volumeMounts[1].SubPath, "known_hosts")
	assert.Equal(t, volumeMounts[1].MountPath, "/root/.ssh/known_hosts")
}

func TestMakeGitVolumesWithKnownHostsFromEnvironment(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.KnownHosts = ``

	os.Setenv("KNOWN_HOSTS_CONFIGMAP", "my-configmap")
	defer os.Unsetenv("KNOWN_HOSTS_CONFIGMAP")

	volumes, volumeMounts := MakeGitVolumes(cr)
	assert.Equal(t, 2, len(volumes))
	assert.Equal(t, volumes[0].Name, "git-key")
	assert.Equal(t, volumes[1].Name, "known-hosts")
	assert.Equal(t, volumes[1].VolumeSource.ConfigMap.LocalObjectReference.Name, KnownHostsName(cr))
	assert.Equal(t, 2, len(volumeMounts))
	assert.Equal(t, volumeMounts[0].Name, "git-key")
	assert.Equal(t, volumeMounts[1].Name, "known-hosts")
	assert.Equal(t, volumeMounts[1].SubPath, "known_hosts")
	assert.Equal(t, volumeMounts[1].MountPath, "/root/.ssh/known_hosts")
}
