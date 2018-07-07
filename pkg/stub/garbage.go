package stub

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
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
		deletePropagation := metav1.DeletePropagationBackground
		err := sdk.Delete(existing, sdk.WithDeleteOptions(&metav1.DeleteOptions{
			PropagationPolicy: &deletePropagation,
		}))
		if err != nil {
			return err
		}
	}

	return nil
}
