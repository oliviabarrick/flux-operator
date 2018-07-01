package stub

import (
	"context"

	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
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

		err = SynchronizeFluxState(o)
	}
	return
}

// Create a flux and tiller with all of the proper RBAC settings.
func SynchronizeFluxState(cr *v1alpha1.Flux) error {
	desiredObjs, err := DesiredFluxObjects(cr)
	if err != nil {
		logrus.Errorf("Failed to determine desired flux state: %v", err)
		return err
	}

	existingObjs, err := ExistingFluxObjects(cr)
	if err != nil {
		logrus.Errorf("Failed to collect existing resources: %v", err)
		return err
	}

	err = CreateOrUpdate(cr, existingObjs, desiredObjs)
	if err != nil {
		logrus.Errorf("Error creating resources: %s", err)
		return err
	}

	err = GarbageCollectResources(cr, existingObjs, desiredObjs)
	if err != nil {
		logrus.Errorf("Error garbage collecting resources: %s", err)
		return err
	}

	return nil
}
