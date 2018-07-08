package fluxcloud

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Generate fluxcloud name
func FluxcloudName(cr *v1alpha1.Flux) string {
	return fmt.Sprintf("flux-%s-fluxcloud", cr.ObjectMeta.Name)
}

// NewFluxcloudService creates a new Fluxcloud service
func NewFluxcloudService(cr *v1alpha1.Flux) *corev1.Service {
	if cr.Spec.FluxCloud.Enabled == false {
		return nil
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, FluxcloudName(cr)),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "fluxcloud",
					Port:       80,
					TargetPort: intstr.FromInt(3031),
				},
			},
			Selector: map[string]string{
				"name": FluxcloudName(cr),
			},
		},
	}
}

// Returns the image name for a fluxcloud instance.
func FluxcloudImage(cr *v1alpha1.Flux) string {
	fluxcloudImage := utils.Getenv("FLUXCLOUD_IMAGE", cr.Spec.FluxCloud.FluxCloudImage)
	if fluxcloudImage == "" {
		fluxcloudImage = utils.FluxcloudImage
	}

	fluxcloudVersion := utils.Getenv("FLUXCLOUD_VERSION", cr.Spec.FluxCloud.FluxCloudVersion)
	if fluxcloudVersion == "" {
		fluxcloudVersion = utils.FluxcloudVersion
	}

	return fmt.Sprintf("%s:%s", fluxcloudImage, fluxcloudVersion)
}

// NewFluxcloudDeployment creates a new fluxcloud deployment
func NewFluxcloudDeployment(cr *v1alpha1.Flux) *extensions.Deployment {
	if cr.Spec.FluxCloud.Enabled == false {
		return nil
	}

	labels := map[string]string{
		"name": FluxcloudName(cr),
	}
	meta := utils.NewObjectMeta(cr, FluxcloudName(cr))
	meta.Labels = labels

	replicas := int32(1)

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
					Containers: []corev1.Container{
						{
							Name:            "fluxcloud",
							Image:           FluxcloudImage(cr),
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 3031,
								},
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "SLACK_URL",
									Value: cr.Spec.FluxCloud.SlackURL,
								},
								corev1.EnvVar{
									Name:  "SLACK_CHANNEL",
									Value: cr.Spec.FluxCloud.SlackChannel,
								},
								corev1.EnvVar{
									Name:  "SLACK_USERNAME",
									Value: cr.Spec.FluxCloud.SlackUsername,
								},
								corev1.EnvVar{
									Name:  "SLACK_ICON_EMOJI",
									Value: cr.Spec.FluxCloud.SlackIconEmoji,
								},
								corev1.EnvVar{
									Name:  "GITHUB_URL",
									Value: cr.Spec.FluxCloud.GithubURL,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU:    resource.MustParse("500m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// Create all of the resources necessary to create a memcached instance.
func NewFluxcloud(cr *v1alpha1.Flux) []runtime.Object {
	if cr.Spec.FluxCloud.Enabled == false {
		return []runtime.Object{}
	}

	return []runtime.Object{NewFluxcloudDeployment(cr), NewFluxcloudService(cr)}
}
