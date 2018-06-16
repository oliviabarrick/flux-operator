package flux

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Create flux command arguments from CR
func MakeFluxArgs(cr *v1alpha1.Flux) (args []string) {
	argMap := map[string]string{
		"git-url": cr.Spec.GitUrl,
		"git-branch": cr.Spec.GitBranch,
		"git-sync-tag": "flux-sync-" + cr.Spec.GitBranch,
		"git-path": cr.Spec.GitPath,
		"git-poll-interval": cr.Spec.GitPollInterval,
	}

	for key, value := range cr.Spec.Args {
		argMap[key] = value
	}

	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	return
}

func NewObjectMeta(cr *v1alpha1.Flux, name string) metav1.ObjectMeta {
	if name == "" {
		name = fmt.Sprintf("flux-%s", cr.Name)
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: cr.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(cr, schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Kind:    "Flux",
			}),
		},
	}
}

// NewFluxPod creates a new flux pod
func NewFluxPod(cr *v1alpha1.Flux) *corev1.Pod {
	serviceAccount := "flux"

	fluxImage := "quay.io/weaveworks/flux"
	if cr.Spec.FluxImage != "" {
		fluxImage = cr.Spec.FluxImage
	}

	fluxVersion := "1.2.3"
	if cr.Spec.FluxVersion != "" {
		fluxVersion = cr.Spec.FluxVersion
	}

	gitSecret := "flux-git-deploy"
	if cr.Spec.GitSecret != "" {
		gitSecret = cr.Spec.GitSecret
  }

	labels := map[string]string{
		"app": "flux",
	}

	meta := NewObjectMeta(cr, "")
	meta.Labels = labels

	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: meta,
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccount,
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "git-key",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: gitSecret,
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
