package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/cnf/structhash"
	"github.com/google/go-github/github"
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strconv"
	"strings"
)

const (
	FLUX_LABEL          = "flux.codesink.net.flux"
	FluxcloudImage      = "justinbarrick/fluxcloud"
	FluxcloudVersion    = "v0.2.1"
	FluxOperatorImage   = "justinbarrick/flux-operator"
	FluxImage           = "quay.io/weaveworks/flux"
	FluxVersion         = "1.8.1"
	HelmOperatorImage   = "quay.io/weaveworks/helm-operator"
	HelmOperatorVersion = "0.4.0"
	MemcachedImage      = "memcached"
	MemcachedVersion    = "1.4.36-alpine"
	TillerImage         = "gcr.io/kubernetes-helm/tiller"
	TillerVersion       = "v2.9.1"
)

// Get an environment variable, pass through strconv.ParseBool, return false if there
// is an error.
func BoolEnv(name string) bool {
	val, _ := strconv.ParseBool(os.Getenv(name))
	return val
}

// Return the namespace the CR's resources should be created in.
func FluxNamespace(cr *v1alpha1.Flux) string {
	if cr.ObjectMeta.Namespace == "" {
		if cr.Spec.Namespace == "" {
			return "default"
		} else {
			return cr.Spec.Namespace
		}
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

// Return the labels that should be set on any object owned by a Flux.
func FluxLabels(cr *v1alpha1.Flux) map[string]string {
	label := cr.ObjectMeta.Name
	if cr.ObjectMeta.Namespace != "" {
		label = fmt.Sprintf("%s-%s", cr.ObjectMeta.Namespace, label)
	}

	return map[string]string{
		FLUX_LABEL: label,
	}
}

// Return the list options that can be used to find an object owned by a Flux.
func ListOptionsForFlux(cr *v1alpha1.Flux) *metav1.ListOptions {
	return &metav1.ListOptions{LabelSelector: labels.SelectorFromSet(FluxLabels(cr)).String()}
}

// Return true if the object is owned by the Flux.
func OwnedByFlux(cr *v1alpha1.Flux, obj runtime.Object) bool {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		fmt.Println(err)
	}

	labels := objectMeta.GetLabels()
	if labels == nil {
		return false
	}

	for k, v := range FluxLabels(cr) {
		if labels[k] != v {
			return false
		}
	}

	return true
}

// Set the owner of an object.
func SetObjectOwner(cr *v1alpha1.Flux, obj runtime.Object) {
	objectMeta, err := meta.Accessor(obj)
	if err != nil {
		fmt.Println(err)
		return
	}

	labels := objectMeta.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	for k, v := range FluxLabels(cr) {
		labels[k] = v
	}

	objectMeta.SetLabels(labels)
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

// Return a human readable name for an object.
func ObjectName(object runtime.Object) string {
	objectMeta, err := meta.Accessor(object)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%s:%s/%s", objectMeta.GetNamespace(),
		object.GetObjectKind().GroupVersionKind().Kind,
		objectMeta.GetName())
}

// Return a human readable string representing the object.
func ReadableObjectName(cr *v1alpha1.Flux, object runtime.Object) string {
	return fmt.Sprintf("flux instance '%s': %s (hash: %s)",
		cr.Name, ObjectName(object), GetObjectHash(object))
}

// Return true if first and second have the same Name, Namespace, and Kind.
func ObjectNameMatches(first runtime.Object, second runtime.Object) bool {
	firstMeta, _ := meta.Accessor(first)
	secondMeta, _ := meta.Accessor(second)

	if firstMeta.GetName() != secondMeta.GetName() {
		return false
	}

	if firstMeta.GetNamespace() != secondMeta.GetNamespace() {
		return false
	}

	if first.GetObjectKind().GroupVersionKind() != second.GetObjectKind().GroupVersionKind() {
		return false
	}

	return true
}

// Return the object from existing that matches Name, Namespace, and Kind with object.
func GetObject(object runtime.Object, existing []runtime.Object) runtime.Object {
	for _, obj := range existing {
		if ObjectNameMatches(obj, object) {
			return obj
		}
	}

	return nil
}

func LatestRelease(repo string) (string, error) {
	client := github.NewClient(nil)

	repo_parts := strings.Split(repo, "/")
	if len(repo_parts) != 2 {
		return "", errors.New(fmt.Sprintf("Could not parse Github repository name: %s", repo))
	}

	release, _, err := client.Repositories.GetLatestRelease(context.TODO(), repo_parts[0], repo_parts[1])
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}
