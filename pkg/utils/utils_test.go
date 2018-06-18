package utils

import (
	"testing"
	"os"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	"github.com/stretchr/testify/assert"
)

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
	assert.Equal(t, objectMeta.Namespace, cr.Spec.Namespace)
	assert.Equal(t, objectMeta.OwnerReferences[0].Kind, "Flux")
}

func TestNewObjectMetaWithName(t *testing.T) {
	assert.Equal(t, NewObjectMeta(test_utils.NewFlux(), "myname").Name, "myname")
}

func TestHashObject(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, HashObject(cr), "6ce0789a57ff8e2b3e487ca910f3d43a09818e63")
	cr.ObjectMeta.Name = "hello"
	assert.Equal(t, HashObject(cr), "4f7153e18b9463f94049e5317975c569ee44fb76")
}

func TestObjectHash(t *testing.T) {
	cr := test_utils.NewFlux()
	assert.Equal(t, GetObjectHash(cr), "")
	SetObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "6ce0789a57ff8e2b3e487ca910f3d43a09818e63")
	ClearObjectHash(cr)
	assert.Equal(t, GetObjectHash(cr), "")
}

func TestClearObjectHashDoesNothingIfNoHashSet(t *testing.T) {
	cr := test_utils.NewFlux()
	ClearObjectHash(cr)
}
