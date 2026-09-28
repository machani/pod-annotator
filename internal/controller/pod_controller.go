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
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	DefaultTargetNamespace  = "maas-t-tenant5"
	DefaultTargetLabelKey   = "prometheus"
	DefaultTargetLabelValue = "maas-t-tenant5-prometheus"
	DefaultAnnotationKey    = "telemetry-compass.com/reconciled"
	DefaultAnnotationValue  = "true"
)

// PodReconciler reconciles Pods that match the configured namespace and label.
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	TargetNamespace  string
	TargetLabelKey   string
	TargetLabelValue string
	AnnotationKey    string
	AnnotationValue  string
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;patch;update

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithValues("pod", req.NamespacedName)

	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("Pod no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if pod.Namespace != r.TargetNamespace {
		log.V(1).Info("Skipping Pod: namespace does not match")
		return ctrl.Result{}, nil
	}

	if pod.Labels[r.TargetLabelKey] != r.TargetLabelValue {
		log.V(1).Info("Skipping Pod: label does not match", "labelKey", r.TargetLabelKey)
		return ctrl.Result{}, nil
	}

	if pod.Annotations != nil && pod.Annotations[r.AnnotationKey] == r.AnnotationValue {
		log.V(1).Info("Pod already has desired annotation")
		return ctrl.Result{}, nil
	}

	before := pod.DeepCopy()
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[r.AnnotationKey] = r.AnnotationValue

	if err := r.Patch(ctx, pod, client.MergeFrom(before)); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Annotated Pod", "annotationKey", r.AnnotationKey, "annotationValue", r.AnnotationValue)
	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	matchesTarget := predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return r.matches(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return r.matches(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return r.matches(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return r.matches(e.Object) },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}, builder.WithPredicates(matchesTarget)).
		Named("pod").
		Complete(r)
}

func (r *PodReconciler) matches(obj client.Object) bool {
	return obj.GetNamespace() == r.TargetNamespace &&
		obj.GetLabels()[r.TargetLabelKey] == r.TargetLabelValue
}
