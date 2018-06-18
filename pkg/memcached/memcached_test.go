package memcached

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/justinbarrick/flux-operator/pkg/utils/test"
	corev1 "k8s.io/api/core/v1"
)

func TestMemcachedName(t *testing.T) {
	cr := test_utils.NewFlux()
	name := MemcachedName(cr)
	assert.Equal(t, name, "flux-" + cr.ObjectMeta.Name + "-memcached")
}

func TestNewMemcachedPod(t *testing.T) {
	cr := test_utils.NewFlux()
	pod := NewMemcachedPod(cr)

	assert.Equal(t, pod.ObjectMeta.Name, MemcachedName(cr))
	assert.Equal(t, pod.ObjectMeta.Namespace, "default")

	c := pod.Spec.Containers[0]
	assert.Equal(t, c.Image, "memcached:1.4.25")
	assert.Equal(t, c.Args, []string{"-m 64", "-p 11211", "-vv"})
}

func TestNewMemcachedService(t *testing.T) {
	cr := test_utils.NewFlux()
	service := NewMemcachedService(cr)

	assert.Equal(t, service.ObjectMeta.Name, MemcachedName(cr))
	assert.Equal(t, service.ObjectMeta.Namespace, "default")

	assert.Equal(t, service.Spec.Ports[0].Name, "memcached")
	assert.Equal(t, service.Spec.Ports[0].Port, int32(11211))
	assert.Equal(t, service.Spec.Selector["name"], MemcachedName(cr))
}

func TestNewMemcached(t *testing.T) {
	cr := test_utils.NewFlux()
	objects := NewMemcached(cr)

	assert.Equal(t, len(objects), 2)
	_ = objects[0].(*corev1.Pod)
	_ = objects[1].(*corev1.Service)
}

