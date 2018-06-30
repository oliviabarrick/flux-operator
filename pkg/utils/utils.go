package utils

import (
	"fmt"
	"os"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/cnf/structhash"
)

// Return the namespace the CR's resources should be created in.
func FluxNamespace(cr *v1alpha1.Flux) string {
	if cr.ObjectMeta.Namespace == "" {
		return cr.Spec.Namespace
	} else {
		return cr.ObjectMeta.Namespace
	}
}

// Getenv returns an environment variable or `value` if it does not exist.
func Getenv(name, value string) string {
	ret := os.Getenv(name)
	if ret == "" {
		return value
	}
	return ret
}

// Returns an ObjectMeta for a CR, if name is an empty string it defaults to
// `"flux-"` + `cr.Name`
func NewObjectMeta(cr *v1alpha1.Flux, name string) metav1.ObjectMeta {
	if name == "" {
		name = fmt.Sprintf("flux-%s", cr.Name)
	}

	return metav1.ObjectMeta{
		Name:      name,
		Namespace: FluxNamespace(cr),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(cr, schema.GroupVersionKind{
				Group:   v1alpha1.SchemeGroupVersion.Group,
				Version: v1alpha1.SchemeGroupVersion.Version,
				Kind:    "Flux",
			}),
		},
	}
}

// Takes a Kubernetes object and returns the hash in its annotations as a string.
func GetObjectHash(obj runtime.Object) string {
		objectMeta, _ := meta.Accessor(obj)

		annotations := objectMeta.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		return annotations["flux.codesink.net.hash"]
}

// Takes a Kubernetes object and adds an annotation with its hash.
func SetObjectHash(obj runtime.Object) {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		fmt.Println(err)
	}

	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["flux.codesink.net.hash"] = HashObject(obj)
	objectMeta.SetAnnotations(annotations)
}

// Takes a Kubernetes object and removes the annotation with its hash.
func ClearObjectHash(obj runtime.Object) {
	objectMeta, _ := meta.Accessor(obj)

	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		return
	}

	delete(annotations, "flux.codesink.net.hash")
	objectMeta.SetAnnotations(annotations)
}

// Return a SHA1 hash of a Kubernetes object
func HashObject(obj runtime.Object) string {
	copied := obj.DeepCopyObject()
	objectMeta, _ := meta.Accessor(copied)

	ownerReferences := objectMeta.GetOwnerReferences()
	if len(ownerReferences) > 0 {
		ownerReferences[0].UID = ""
	}
	objectMeta.SetOwnerReferences(ownerReferences)

	return fmt.Sprintf("%x", structhash.Sha1(copied, 1))
}

// Return a human readable string representing the object.
func ReadableObjectName(cr *v1alpha1.Flux, object runtime.Object) string {
	objectMeta, err := meta.Accessor(object)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("flux instance '%s': %s %s/%s (hash: %s)",
		cr.Name, object.GetObjectKind().GroupVersionKind().Kind,
		objectMeta.GetNamespace(), objectMeta.GetName(), GetObjectHash(object))
}
