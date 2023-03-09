/*
Copyright 2023 zhiwei@youya.org.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validator

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:webhook:path=/validate-v1-persistentvolumeclaim,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create,versions=v1,name=vpersistentvolumeclaim.kb.io

// PersistentVolumeClaimValidator validates PVCs
type PersistentVolumeClaimValidator struct {
	Client client.Client
}

// validate admits a pvc if a storageclass is allowed to use
func (v *PersistentVolumeClaimValidator) validate(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)
	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("expected a PersistentVolumeClaim but got a %T", obj)
	}

	log.Info("Validating PersistentVolumeClaim")

	ns := pvc.GetNamespace()
	sc := pvc.Spec.StorageClassName

	nsSA := "system:serviceaccounts:" + ns

	sar := &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			Groups: []string{nsSA},
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:     "storage.k8s.io",
				Resource:  "storageclasses",
				Name:      *sc,
				Namespace: ns,
				Verb:      "use",
			},
		},
	}

	if err := v.Client.Create(ctx, sar); err != nil {
		return err
	}

	if sar.Status.Allowed {
		return nil
	} else {
		return fmt.Errorf("create pvc with storageclass %s in namespace %s is not allowed", *sc, ns)
	}
}

func (v *PersistentVolumeClaimValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return v.validate(ctx, obj)
}

func (v *PersistentVolumeClaimValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	return nil
}

func (v *PersistentVolumeClaimValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}
