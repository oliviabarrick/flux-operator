package stub

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/flux"
	"github.com/justinbarrick/flux-operator/pkg/fluxcloud"
	"github.com/justinbarrick/flux-operator/pkg/helm-operator"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/tiller"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

// Create flux, tiller, and helm-operator instances from a CR and return them
// as a list of objects.
func DesiredFluxObjects(cr *v1alpha1.Flux) ([]runtime.Object, error) {
	objects := rbac.FluxRoles(cr)
	dep := flux.NewFluxDeployment(cr)
	objects = append(objects, dep)
	objects = append(objects, memcached.NewMemcached(cr)...)
	objects = append(objects, fluxcloud.NewFluxcloud(cr)...)

	sshKey := flux.NewFluxSSHKey(cr)
	err := sdk.Get(sshKey)
	if err != nil || utils.OwnedByFlux(cr, sshKey) {
		objects = append(objects, flux.NewFluxSSHKey(cr))
	}

	knownHosts := flux.NewFluxKnownHosts(cr)
	if knownHosts != nil {
		objects = append(objects, knownHosts)
	}

	tillerObjects, err := tiller.NewTiller(cr)
	if err != nil {
		logrus.Errorf("Failed to create tiller instance: %v", err)
		return nil, err
	}

	objects = append(objects, tillerObjects...)

	helmOperator := helm_operator.NewHelmOperatorDeployment(cr)
	if helmOperator != nil {
		objects = append(objects, helmOperator)
	}

	for index, object := range objects {
		utils.SetObjectOwner(cr, object)
		utils.SetObjectHash(object)
		objects[index] = object
	}

	return objects, nil
}
