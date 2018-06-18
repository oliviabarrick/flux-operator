package flux

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/justinbarrick/flux-operator/pkg/utils"
)

func GitSecretName(cr *v1alpha1.Flux) string {
	secretName := utils.Getenv("GIT_SECRET_NAME", fmt.Sprintf("flux-git-%s-deploy", cr.Name))

	if cr.Spec.GitSecret != "" {
		secretName = cr.Spec.GitSecret
  }

	return secretName
}

// Create flux command arguments from CR
func MakeFluxArgs(cr *v1alpha1.Flux) (args []string) {
	branch := cr.Spec.GitBranch
	if branch == "" {
		branch = "master"
	}

	path := cr.Spec.GitPath
	if path == "" {
		path = "./"
	}

	poll := cr.Spec.GitPollInterval
	if poll == "" {
		poll = "5m30s"
	}

	argMap := map[string]string{
		"git-url": cr.Spec.GitUrl,
		"git-branch": branch,
		"git-sync-tag": fmt.Sprintf("flux-sync-%s", cr.Name),
		"git-path": path,
		"git-poll-interval": poll,
		"k8s-secret-name": GitSecretName(cr),
		"memcached-hostname": memcached.MemcachedName(cr),
	}

	for key, value := range cr.Spec.Args {
		argMap[key] = value
	}

	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	return
}

// NewFluxPod creates a new flux pod
func NewFluxPod(cr *v1alpha1.Flux) *corev1.Pod {
	fluxImage := utils.Getenv("FLUX_IMAGE", "quay.io/weaveworks/flux")
	if cr.Spec.FluxImage != "" {
		fluxImage = cr.Spec.FluxImage
	}

	fluxVersion := utils.Getenv("FLUX_VERSION", "1.4.0")
	if cr.Spec.FluxVersion != "" {
		fluxVersion = cr.Spec.FluxVersion
	}

	labels := map[string]string{
		"app": "flux",
	}

	meta := utils.NewObjectMeta(cr, "")
	meta.Labels = labels

	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec: corev1.PodSpec{
			ServiceAccountName: rbac.ServiceAccountName(cr),
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "git-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: GitSecretName(cr),
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "flux",
					Image:   fmt.Sprintf("%s:%s", fluxImage, fluxVersion),
					ImagePullPolicy: "IfNotPresent",
					Ports: []corev1.ContainerPort{
						corev1.ContainerPort{
							ContainerPort: 3030,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name: "git-key",
							MountPath: "/etc/fluxd/ssh",
						},
					},
					Args: MakeFluxArgs(cr),
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("512Mi"),
							corev1.ResourceCPU: resource.MustParse("500m"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("256Mi"),
							corev1.ResourceCPU: resource.MustParse("500m"),
						},
					},
				},
			},
		},
	}
}

func NewFluxSSHKey(cr *v1alpha1.Flux) *corev1.Secret {
	return &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, GitSecretName(cr)),
		Type: "opaque",
	}
}
