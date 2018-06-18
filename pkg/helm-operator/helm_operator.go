package helm_operator

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/flux"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/justinbarrick/flux-operator/pkg/utils"
)

// Create helm-operator command arguments from CR
func MakeHelmOperatorArgs(cr *v1alpha1.Flux) (args []string) {
	branch := cr.Spec.GitBranch
	if branch == "" {
		branch = "master"
	}

	path := cr.Spec.HelmOperator.ChartPath
	if path == "" {
		path = "./"
	}

	poll := cr.Spec.HelmOperator.GitPollInterval
	if poll == "" {
		poll = cr.Spec.GitPollInterval

		if poll == "" {
			poll = "3m0s"
		}
	}

	gitUrl := cr.Spec.HelmOperator.GitUrl
	if gitUrl == "" {
		gitUrl = cr.Spec.GitUrl
	}

	argMap := map[string]string{
		"git-url": gitUrl,
		"git-branch": branch,
		"git-charts-path": path,
		"charts-sync-interval": poll,
		"tiller-namespace": cr.Spec.Namespace,

	}

	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	return
}

// NewHelmOperatorPod creates a new helm-operator pod
func NewHelmOperatorPod(cr *v1alpha1.Flux) *corev1.Pod {
	if ! cr.Spec.HelmOperator.Enabled {
		return nil
	}

	operatorImage := utils.Getenv("HELM_OPERATOR_IMAGE", "quay.io/weaveworks/helm-operator")
	if cr.Spec.HelmOperator.HelmOperatorImage != "" {
		operatorImage = cr.Spec.HelmOperator.HelmOperatorImage
	}

	operatorVersion := utils.Getenv("HELM_OPERATOR_VERSION", "master-1dfdc61")
	if cr.Spec.HelmOperator.HelmOperatorVersion != "" {
		operatorVersion = cr.Spec.HelmOperator.HelmOperatorVersion
	}

	labels := map[string]string{
		"app": "helm-operator",
	}

	meta := utils.NewObjectMeta(cr, fmt.Sprintf("helm-operator-%s", cr.ObjectMeta.Name))
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
							SecretName: flux.GitSecretName(cr),
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:    "helm-operator",
					Image:   fmt.Sprintf("%s:%s", operatorImage, operatorVersion),
					ImagePullPolicy: "IfNotPresent",
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name: "git-key",
							MountPath: "/etc/fluxd/ssh",
						},
					},
					Args: MakeHelmOperatorArgs(cr),
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
