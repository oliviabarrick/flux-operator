package utils

import (
	"testing"
	"os"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
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

func TestGetenv(t *testing.T) {
	os.Setenv("MY_VAR", "value")
	defer os.Setenv("MY_VAR", "")
	assert.Equal(t, Getenv("MY_VAR", "othervalue"), "value")
	assert.Equal(t, Getenv("NON_EXISTANT_ENV_VAR", "MYVALUE"), "MYVALUE")
}

func TestNewObjectMeta(t *testing.T) {
	cr := test_utils.NewFlux()
	objectMeta := NewObjectMeta(cr, "")
	assert.Equal(t, objectMeta.Name, "flux-" + cr.Name)
	assert.Equal(t, objectMeta.Namespace, FluxNamespace(cr))
	assert.Equal(t, objectMeta.OwnerReferences[0].Kind, "Flux")
}

func TestNewObjectMetaWithName(t *testing.T) {
	assert.Equal(t, NewObjectMeta(test_utils.NewFlux(), "myname").Name, "myname")
}

func TestHashObject(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, HashObject(cr), "f3c2a42e485dadf412f495ae5e5bcf7b90bb7349")
	cr.ObjectMeta.Name = "hello"
	assert.Equal(t, HashObject(cr), "fa5d630d05d2c6974cf99a8e06e76f4b09e3f407")
}

func TestObjectHash(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, GetObjectHash(cr), "")
	SetObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "f3c2a42e485dadf412f495ae5e5bcf7b90bb7349")
	ClearObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "")
}

func TestClearObjectHashDoesNothingIfNoHashSet(t *testing.T) {
	cr := test_utils.NewFlux()
	ClearObjectHash(cr)
}
