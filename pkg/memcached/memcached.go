package memcached

import (
	"fmt"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/justinbarrick/flux-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
)

// Generate memcached name
func MemcachedName(cr *v1alpha1.Flux) string {
	return fmt.Sprintf("flux-%s-memcached", cr.ObjectMeta.Name)
}

// NewMemcachedService creates a new memcached service
func NewMemcachedService(cr *v1alpha1.Flux) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: utils.NewObjectMeta(cr, MemcachedName(cr)),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "memcached",
					Port: 11211,
				},
			},
			Selector: map[string]string{
				"name": MemcachedName(cr),
			},
		},
	}
}

// NewMemcachedDeployment creates a new memcached deployment
func NewMemcachedDeployment(cr *v1alpha1.Flux) *extensions.Deployment {
	memcachedImage := utils.Getenv("MEMCACHED_IMAGE", "memcached")
	memcachedVersion := utils.Getenv("MEMCACHED_VERSION", "1.4.25")

	labels := map[string]string{
		"name": MemcachedName(cr),
	}
	meta := utils.NewObjectMeta(cr, MemcachedName(cr))
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
							Name:    "memcached",
							Image:   fmt.Sprintf("%s:%s", memcachedImage, memcachedVersion),
							ImagePullPolicy: "IfNotPresent",
							Args: []string{"-m 64", "-p 11211", "-vv"},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 11211,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU: resource.MustParse("500m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU: resource.MustParse("100m"),
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
func NewMemcached(cr *v1alpha1.Flux) []runtime.Object {
	return []runtime.Object{NewMemcachedDeployment(cr), NewMemcachedService(cr)}
}
