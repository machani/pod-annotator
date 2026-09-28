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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pod Controller", func() {
	Context("When reconciling a resource", func() {

		It("should annotate a matching pod and preserve existing annotations", func() {
			By("creating the target namespace")

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: targetNamespace,
				},
			}

			Expect(k8sClient.Create(ctx, namespace)).To(Succeed())

			By("creating a pod that matches the namespace and label")

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "matching-pod",
					Namespace: targetNamespace,
					Labels: map[string]string{
						targetLabelKey: targetLabelValue,
					},
					Annotations: map[string]string{
						"owner": "machani",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			// ...rest of test unchanged
		})
	})
})
