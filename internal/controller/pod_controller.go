/*
Copyright 2026.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	targetNamespace  = "maas-t-tenant5"
	targetLabelKey   = "prometheus"
	targetLabelValue = "maas-t-tenant5-prometheus"

	annotationKey   = "telemetry-compass.com/reconciled"
	annotationValue = "true"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.25.0/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	pod := &corev1.Pod{}

	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Pod not found, probably deleted")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	// Check namespace.
	if pod.Namespace != targetNamespace {
		log.Info("Skipping Pod: namespace does not match",
			"name", pod.Name,
			"namespace", pod.Namespace,
		)
		return ctrl.Result{}, nil
	}

	// Check label.
	labelValue, exists := pod.Labels[targetLabelKey]
	if !exists || labelValue != targetLabelValue {
		log.Info("Skipping Pod: label does not match",
			"name", pod.Name,
			"namespace", pod.Namespace,
		)
		return ctrl.Result{}, nil
	}

	// Don't update the Pod if it already has the desired annotation.
	if pod.Annotations != nil && pod.Annotations[annotationKey] == annotationValue {
		log.Info("Pod already has annotation",
			"name", pod.Name,
			"namespace", pod.Namespace,
		)
		return ctrl.Result{}, nil
	}

	// A Pod might not have an annotations map yet.
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	// Modify our local Pod object.
	pod.Annotations[annotationKey] = annotationValue

	// Persist the change to Kubernetes.
	if err := r.Update(ctx, pod); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Added annotation to Pod",
		"name", pod.Name,
		"namespace", pod.Namespace,
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Named("pod").
		Complete(r)
}
