package stub

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// Remove any resources that should no longer exist.
// If a CR is completed deleted, it will be deleted automatically by Kubernetes finalizers.
// This method deletes resources that need to be deleted if a CR is updated and a
// resource is obsolete, e.g., helmOperator.enabled is changed from true to false.
func GarbageCollectResources(cr *v1alpha1.Flux, existingObjs []runtime.Object, desiredObjs []runtime.Object) error {
	for _, existing := range existingObjs {
		desired := utils.GetObject(existing, desiredObjs)
		if desired != nil {
			continue
		}

		logrus.Infof("Deleting unwanted resource from %s", utils.ReadableObjectName(cr, existing))
		err := sdk.Delete(existing)
		if err != nil {
			return err
		}
	}

	return nil
}
