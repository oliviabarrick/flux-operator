package stub

import (
	"context"

	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/flux"
	"github.com/justinbarrick/flux-operator/pkg/helm-operator"
	"github.com/justinbarrick/flux-operator/pkg/memcached"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/tiller"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/api/meta"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) (err error) {
	switch o := event.Object.(type) {
	case *v1alpha1.Flux:
		if event.Deleted {
			return
		}

		err = CreateFlux(o)
	}
	return
}

// Create flux, tiller, and helm-operator instances from a CR and return them
// as a list of objects.
func CreateFluxObjects(cr *v1alpha1.Flux) ([]runtime.Object, error) {
	objects := rbac.FluxRoles(cr)
	dep := flux.NewFluxDeployment(cr)
	objects = append(objects, dep)
	objects = append(objects, memcached.NewMemcached(cr)...)

	err := sdk.Get(flux.NewFluxSSHKey(cr))
	if err != nil {
		objects = append(objects, flux.NewFluxSSHKey(cr))
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
		utils.SetObjectHash(object)
		objects[index] = object
	}

	return objects, nil
}

// Create a flux and tiller with all of the proper RBAC settings.
func CreateFlux (cr *v1alpha1.Flux) error {
	objects, err := CreateFluxObjects(cr)
	if err != nil {
		logrus.Errorf("Failed to create flux instance: %v", err)
		return err
	}

	for _, object := range objects {
		name := utils.ReadableObjectName(cr, object)

		if err != nil {
			logrus.Errorf("Could not generate object name: %v", err)
		}

		err = sdk.Create(object)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create %s: %v", name, err)
			return err
		} else if err != nil && errors.IsAlreadyExists(err) {
			oldObject := object.DeepCopyObject()

			utils.ClearObjectHash(oldObject)
			err = sdk.Get(oldObject)
			if err != nil {
				return err
			}

			if utils.GetObjectHash(oldObject) == utils.GetObjectHash(object) {
				continue
			}

			switch object.(type) {
				case *corev1.Service:
					service := object.(*corev1.Service)
					service.Spec.ClusterIP = oldObject.(*corev1.Service).Spec.ClusterIP
			}

			oldObjectMeta, _ := meta.Accessor(oldObject)
			newObjectMeta, _ := meta.Accessor(object)
			newObjectMeta.SetResourceVersion(oldObjectMeta.GetResourceVersion())

			err = sdk.Update(object)
			if err != nil {
				logrus.Errorf("Could not update %s", name)
				return err
			}

			logrus.Infof("Updated out of date %s != %s", name, utils.GetObjectHash(oldObject))
		} else {
			logrus.Infof("Created %s", name)
		}
	}

	return nil
}
