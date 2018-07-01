package stub

import (
	"github.com/justinbarrick/flux-operator/pkg/apis/flux/v1alpha1"
	"github.com/justinbarrick/flux-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/api/meta"
	"github.com/sirupsen/logrus"
)

// Create or update desiredObjs based on the current state in existingObjs
func CreateOrUpdate(cr *v1alpha1.Flux, existingObjs []runtime.Object, desiredObjs []runtime.Object) error {
	for _, desired := range desiredObjs {
		name := utils.ReadableObjectName(cr, desired)

		existing := utils.GetObject(desired, existingObjs)
		if existing == nil {
			err := sdk.Create(desired)
			if err != nil {
				logrus.Errorf("Failed to create %s", name)
				return err
			}

			logrus.Infof("Created %s", name)
			continue
		}

		if utils.GetObjectHash(existing) == utils.GetObjectHash(desired) {
			continue
		}

		switch desired.(type) {
			case *corev1.Service:
				service := desired.(*corev1.Service)
				service.Spec.ClusterIP = existing.(*corev1.Service).Spec.ClusterIP
		}

		existingMeta, _ := meta.Accessor(existing)
		desiredMeta, _ := meta.Accessor(desired)
		desiredMeta.SetResourceVersion(existingMeta.GetResourceVersion())

		err := sdk.Update(desired)
		if err != nil {
			logrus.Errorf("Could not update %s", name)
			return err
		}

		logrus.Infof("Updated out of date %s != %s", name, utils.GetObjectHash(existing))
	}

	return nil
}
