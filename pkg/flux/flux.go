package flux

import (
	"fmt"
	"sort"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extensions "k8s.io/api/extensions/v1beta1"
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
		"ssh-keygen-dir": "/etc/fluxd/",
		"memcached-hostname": memcached.MemcachedName(cr),
	}

	for key, value := range cr.Spec.Args {
		argMap[key] = value
	}

	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	sort.Strings(args)

	return
}

// NewFluxDeployment creates a new flux pod
func NewFluxDeployment(cr *v1alpha1.Flux) *extensions.Deployment {
	fluxImage := utils.Getenv("FLUX_IMAGE", utils.FluxImage)
	if cr.Spec.FluxImage != "" {
		fluxImage = cr.Spec.FluxImage
	}

	fluxVersion := utils.Getenv("FLUX_VERSION", utils.FluxVersion)
	if cr.Spec.FluxVersion != "" {
		fluxVersion = cr.Spec.FluxVersion
	}

	meta := utils.NewObjectMeta(cr, "")
	labels := map[string]string{
		"app": "flux",
		"flux": cr.Name,
	}

	meta.Labels = labels

	replicas := int32(1)
	secretMode := int32(0400)

	return &extensions.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: meta,
		Spec: extensions.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: rbac.ServiceAccountName(cr),
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "git-key",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: GitSecretName(cr),
									DefaultMode: &secretMode,
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
									ReadOnly: true,
								},
							},
							Args: MakeFluxArgs(cr),
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU: resource.MustParse("500m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU: resource.MustParse("250m"),
								},
							},
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
