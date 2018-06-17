package stub

import (
	"context"

	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/flux"

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

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Flux:
		objects := flux.FluxRoles(o)
		objects = append(objects, flux.NewFluxPod(o))

		for _, object := range objects {
			err := sdk.Create(object)
			if err != nil && !errors.IsAlreadyExists(err) {
				logrus.Errorf("Failed to create flux instance: %v", err)
				return err
			}
		}
	}
	return nil
}
