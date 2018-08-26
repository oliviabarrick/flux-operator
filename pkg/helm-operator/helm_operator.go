package helm_operator

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/flux"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/tiller"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
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

	sync := cr.Spec.HelmOperator.ChartsSyncInterval
	if sync == "" {
		sync = cr.Spec.SyncInterval

		if sync == "" {
			sync = "3m0s"
		}
	}

	gitUrl := cr.Spec.HelmOperator.GitUrl
	if gitUrl == "" {
		gitUrl = cr.Spec.GitUrl
	}

	argMap := map[string]string{
		"git-url":              gitUrl,
		"git-branch":           branch,
		"git-charts-path":      path,
		"git-poll-interval":    poll,
		"charts-sync-interval": sync,
		"tiller-namespace":     utils.FluxNamespace(cr),
		"tiller-ip":            tiller.TillerName(cr),
		"tiller-port":          "44134",
	}

	for key, value := range argMap {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	sort.Strings(args)

	return
}

// NewHelmOperatorDeployment creates a new helm-operator deployment
func NewHelmOperatorDeployment(cr *v1alpha1.Flux) *extensions.Deployment {
	if !cr.Spec.HelmOperator.Enabled {
		return nil
	}

	operatorImage := utils.Getenv("HELM_OPERATOR_IMAGE", utils.HelmOperatorImage)
	if cr.Spec.HelmOperator.HelmOperatorImage != "" {
		operatorImage = cr.Spec.HelmOperator.HelmOperatorImage
	}

	operatorVersion := utils.Getenv("HELM_OPERATOR_VERSION", utils.HelmOperatorVersion)
	if cr.Spec.HelmOperator.HelmOperatorVersion != "" {
		operatorVersion = cr.Spec.HelmOperator.HelmOperatorVersion
	}

	labels := map[string]string{
		"app":  "helm-operator",
		"flux": cr.Name,
	}

	meta := utils.NewObjectMeta(cr, fmt.Sprintf("flux-%s-helm-operator", cr.ObjectMeta.Name))
	meta.Labels = labels

	replicas := int32(1)
	secretMode := int32(0400)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("512Mi"),
			corev1.ResourceCPU:    resource.MustParse("1000m"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
			corev1.ResourceCPU:    resource.MustParse("250m"),
		},
	}
	if cr.Spec.HelmOperator.Resources != nil {
		resourceRequirements = *cr.Spec.HelmOperator.Resources
	}

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
									SecretName:  flux.GitSecretName(cr),
									DefaultMode: &secretMode,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "helm-operator",
							Image:           fmt.Sprintf("%s:%s", operatorImage, operatorVersion),
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "git-key",
									MountPath: "/etc/fluxd/ssh",
									ReadOnly:  true,
								},
							},
							Args:      MakeHelmOperatorArgs(cr),
							Resources: resourceRequirements,
						},
					},
				},
			},
		},
	}
}
