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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pod Controller", func() {
	const otherNamespace = "pod-annotator-other"
	var reconciler *PodReconciler

	ensureNamespace := func(name string) {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
		err := k8sClient.Create(ctx, ns)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	}

	newPod := func(name, namespace, labelValue string, annotations map[string]string) *corev1.Pod {
		labels := map[string]string{}
		if labelValue != "" {
			labels[DefaultTargetLabelKey] = labelValue
		}
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels, Annotations: annotations},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "test", Image: "nginx"}}},
		}
	}

	reconcile := func(pod *corev1.Pod) {
		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}})
		Expect(err).NotTo(HaveOccurred())
	}

	getPod := func(pod *corev1.Pod) *corev1.Pod {
		current := &corev1.Pod{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, current)).To(Succeed())
		return current
	}

	BeforeEach(func() {
		ensureNamespace(DefaultTargetNamespace)
		ensureNamespace(otherNamespace)
		reconciler = &PodReconciler{
			Client: k8sClient, Scheme: k8sClient.Scheme(),
			TargetNamespace: DefaultTargetNamespace, TargetLabelKey: DefaultTargetLabelKey,
			TargetLabelValue: DefaultTargetLabelValue, AnnotationKey: DefaultAnnotationKey,
			AnnotationValue: DefaultAnnotationValue,
		}
	})

	It("annotates a matching Pod and preserves unrelated annotations", func() {
		pod := newPod("matching-pod", DefaultTargetNamespace, DefaultTargetLabelValue, map[string]string{"owner": "machani"})
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		reconcile(pod)
		current := getPod(pod)
		Expect(current.Annotations).To(HaveKeyWithValue(DefaultAnnotationKey, DefaultAnnotationValue))
		Expect(current.Annotations).To(HaveKeyWithValue("owner", "machani"))
	})

	It("does not annotate a Pod with a wrong label value", func() {
		pod := newPod("wrong-label-pod", DefaultTargetNamespace, "something-else", nil)
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		reconcile(pod)
		Expect(getPod(pod).Annotations).NotTo(HaveKey(DefaultAnnotationKey))
	})

	It("does not annotate a Pod with the target label missing", func() {
		pod := newPod("missing-label-pod", DefaultTargetNamespace, "", nil)
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		reconcile(pod)
		Expect(getPod(pod).Annotations).NotTo(HaveKey(DefaultAnnotationKey))
	})

	It("does not annotate a Pod in a different namespace", func() {
		pod := newPod("wrong-namespace-pod", otherNamespace, DefaultTargetLabelValue, nil)
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		reconcile(pod)
		Expect(getPod(pod).Annotations).NotTo(HaveKey(DefaultAnnotationKey))
	})

	It("is idempotent when the desired annotation already exists", func() {
		pod := newPod("already-annotated-pod", DefaultTargetNamespace, DefaultTargetLabelValue, map[string]string{DefaultAnnotationKey: DefaultAnnotationValue})
		Expect(k8sClient.Create(ctx, pod)).To(Succeed())
		before := getPod(pod).ResourceVersion
		reconcile(pod)
		current := getPod(pod)
		Expect(current.Annotations).To(HaveKeyWithValue(DefaultAnnotationKey, DefaultAnnotationValue))
		Expect(current.ResourceVersion).To(Equal(before))
	})

	It("returns successfully when the Pod no longer exists", func() {
		_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "deleted-pod", Namespace: DefaultTargetNamespace}})
		Expect(err).NotTo(HaveOccurred())
	})
})
