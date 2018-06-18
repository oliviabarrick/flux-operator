package stub

import (
	"context"

	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/flux"
	"github.com/justinbarrick/flux-operator/pkg/rbac"
	"github.com/justinbarrick/flux-operator/pkg/tiller"
	"github.com/justinbarrick/flux-operator/pkg/helm-operator"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
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

// Create a flux and tiller with all of the proper RBAC settings.
func CreateFlux (cr *v1alpha1.Flux) error {
	objects := rbac.FluxRoles(cr)
	objects = append(objects, flux.NewFluxPod(cr))

	err := sdk.Get(flux.NewFluxSSHKey(cr))
	if err != nil {
		objects = append(objects, flux.NewFluxSSHKey(cr))
	}

	tillerObjects, err := tiller.NewTiller(cr)
	if err != nil {
		logrus.Errorf("Failed to create tiller instance: %v", err)
		return err
	}

	objects = append(objects, tillerObjects...)

	helmOperator := helm_operator.NewHelmOperatorPod(cr)
	if helmOperator != nil {
		objects = append(objects, helmOperator)
	}

	for _, object := range objects {
		err := sdk.Create(object)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create flux instance: %v", err)
			return err
		}
	}

	return nil
}
