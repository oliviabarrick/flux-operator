package utils

import (
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"testing"
)

func TestFluxNamespace(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Namespace = "hello"
	cr.ObjectMeta.Namespace = "mynamespace"
	assert.Equal(t, "mynamespace", FluxNamespace(cr))
}

func TestFluxNamespaceFromSpecIfClusterScope(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Namespace = "hello"
	cr.ObjectMeta.Namespace = ""
	assert.Equal(t, "hello", FluxNamespace(cr))
}

func TestFluxNamespaceFromSpecIfClusterScopeDefault(t *testing.T) {
	cr := test_utils.NewFlux()
	cr.Spec.Namespace = ""
	cr.ObjectMeta.Namespace = ""
	assert.Equal(t, "default", FluxNamespace(cr))
}

func TestGetenv(t *testing.T) {
	os.Setenv("MY_VAR", "value")
	defer os.Setenv("MY_VAR", "")
	assert.Equal(t, Getenv("MY_VAR", "othervalue"), "value")
	assert.Equal(t, Getenv("NON_EXISTANT_ENV_VAR", "MYVALUE"), "MYVALUE")
}

func TestNewObjectMeta(t *testing.T) {
	cr := test_utils.NewFlux()
	objectMeta := NewObjectMeta(cr, "")
	assert.Equal(t, objectMeta.Name, "flux-"+cr.Name)
	assert.Equal(t, objectMeta.Namespace, FluxNamespace(cr))
	assert.Equal(t, objectMeta.OwnerReferences[0].Kind, "Flux")
}

func TestNewObjectMetaWithName(t *testing.T) {
	assert.Equal(t, NewObjectMeta(test_utils.NewFlux(), "myname").Name, "myname")
}

func TestHashObject(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, HashObject(cr), "28c6a21ca7f681a9766df54af628c233a58a1e97")
	cr.ObjectMeta.Name = "hello"
	assert.Equal(t, HashObject(cr), "f73ff9ce9b9efed41f1210764eb57e6a31b3d418")
}

func TestObjectHash(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, GetObjectHash(cr), "")
	SetObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "28c6a21ca7f681a9766df54af628c233a58a1e97")
	ClearObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "")
}

func TestClearObjectHashDoesNothingIfNoHashSet(t *testing.T) {
	cr := test_utils.NewFlux()
	ClearObjectHash(cr)
}

func TestBoolEnv(t *testing.T) {
	assert.Equal(t, BoolEnv("MY_NON_EXISTANT_VAR"), false)
	os.Setenv("MY_VAR", "myval")
	assert.Equal(t, BoolEnv("MY_VAR"), false)
	os.Setenv("MY_VAR", "true")
	assert.Equal(t, BoolEnv("MY_VAR"), true)
	os.Setenv("MY_VAR", "1")
	assert.Equal(t, BoolEnv("MY_VAR"), true)
	os.Setenv("MY_VAR", "false")
	assert.Equal(t, BoolEnv("MY_VAR"), false)
	os.Setenv("MY_VAR", "0")
	assert.Equal(t, BoolEnv("MY_VAR"), false)
	os.Setenv("MY_VAR", "TRUE")
	assert.Equal(t, BoolEnv("MY_VAR"), true)
}

func TestObjectNameMatches(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, ObjectNameMatches(cr, cr), true)

	cr2 := test_utils.NewFlux()
	assert.Equal(t, ObjectNameMatches(cr, cr2), true)

	cr2 = test_utils.NewFlux()
	cr2.ObjectMeta.Name = "newname"
	assert.Equal(t, ObjectNameMatches(cr, cr2), false)

	cr2 = test_utils.NewFlux()
	cr2.ObjectMeta.Namespace = "newnamespace"
	assert.Equal(t, ObjectNameMatches(cr, cr2), false)

	cr2 = test_utils.NewFlux()
	cr2.TypeMeta.Kind = "Hello"
	assert.Equal(t, ObjectNameMatches(cr, cr2), false)

	cr2 = test_utils.NewFlux()
	cr2.ObjectMeta.Annotations = map[string]string{"myannotation": "myannotation"}
	assert.Equal(t, ObjectNameMatches(cr, cr2), true)
}

func TestGetObject(t *testing.T) {
	cr := test_utils.NewFlux()
	cr2 := test_utils.NewFlux()
	cr2.ObjectMeta.Annotations = map[string]string{"myannotation": "myannotation"}

	cr3 := test_utils.NewFlux()
	cr3.ObjectMeta.Name = "newname"

	cr4 := test_utils.NewFlux()
	cr4.ObjectMeta.Name = "othername"

	assert.Equal(t, GetObject(cr, []runtime.Object{cr3, cr2, cr4}), cr2)
	assert.Equal(t, GetObject(cr, []runtime.Object{cr3, cr4}), nil)
}
