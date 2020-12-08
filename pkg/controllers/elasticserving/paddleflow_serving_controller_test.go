/*


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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

var _ = Context("Inside of a new namespace", func() {
	ctx := context.TODO()
	ns := SetupTest(ctx)

	Describe("when no existing resources exist", func() {

		It("should create a new Deployment resource with the specified name and one replica if none is provided", func() {
			paddle := &elasticservingv1.Paddle{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testresource",
					Namespace: ns.Name,
				},
				Spec: elasticservingv1.PaddleSpec{
					DeploymentName: "deployment-name",
					RuntimeVersion: "paddle",
					StorageURI:     "hub.baidubce.com/paddlepaddle/serving:latest",
					Port:           9292,
				},
			}

			err := k8sClient.Create(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Paddle resource")

			deployment := apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddle.Spec.DeploymentName,
					Namespace: paddle.Namespace,
				},
				Spec: apps.DeploymentSpec{
					Replicas: paddle.Spec.Replicas,
					Template: core.PodTemplateSpec{
						Spec: core.PodSpec{
							Containers: []core.Container{
								{
									Name:  paddle.Spec.RuntimeVersion,
									Image: paddle.Spec.StorageURI,
									Ports: []core.ContainerPort{
										{ContainerPort: paddle.Spec.Port, Name: "http", Protocol: "TCP"},
									},
								},
							},
						},
					},
				},
			}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: "deployment-name", Namespace: paddle.Namespace}, &deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
		})

		It("should create a new Deployment resource with the specified name and two replicas if two is specified", func() {
			paddle := &elasticservingv1.Paddle{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testresource",
					Namespace: ns.Name,
				},
				Spec: elasticservingv1.PaddleSpec{
					DeploymentName: "deployment-name",
					Replicas:       pointer.Int32Ptr(2),
					RuntimeVersion: "paddle",
					StorageURI:     "hub.baidubce.com/paddlepaddle/serving:latest",
					Port:           9292,
				},
			}

			err := k8sClient.Create(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Paddle resource")

			deployment := &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddle.Spec.DeploymentName,
					Namespace: paddle.Namespace,
				},
				Spec: apps.DeploymentSpec{
					Replicas: paddle.Spec.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
						},
					},
					Template: core.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
							},
						},
						Spec: core.PodSpec{
							Containers: []core.Container{
								{
									Name:  paddle.Spec.RuntimeVersion,
									Image: paddle.Spec.StorageURI,
									Ports: []core.ContainerPort{
										{ContainerPort: paddle.Spec.Port, Name: "http", Protocol: "TCP"},
									},
								},
							},
						},
					},
				},
			}
			Eventually(
				getResourceFunc(ctx, client.ObjectKey{Name: "deployment-name", Namespace: paddle.Namespace}, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(2)))
		})

		It("should allow updating the replicas count after creating a Paddle resource", func() {
			deploymentObjectKey := client.ObjectKey{
				Name:      "deployment-name",
				Namespace: ns.Name,
			}
			paddleObjectKey := client.ObjectKey{
				Name:      "testresource",
				Namespace: ns.Name,
			}
			paddle := &elasticservingv1.Paddle{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddleObjectKey.Name,
					Namespace: paddleObjectKey.Namespace,
				},
				Spec: elasticservingv1.PaddleSpec{
					DeploymentName: deploymentObjectKey.Name,
					RuntimeVersion: "paddle",
					StorageURI:     "hub.baidubce.com/paddlepaddle/serving:latest",
					Port:           9292,
				},
			}

			err := k8sClient.Create(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Paddle resource")

			deployment := &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddle.Spec.DeploymentName,
					Namespace: paddle.Namespace,
				},
				Spec: apps.DeploymentSpec{
					Replicas: paddle.Spec.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
						},
					},
					Template: core.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
							},
						},
						Spec: core.PodSpec{
							Containers: []core.Container{
								{
									Name:  paddle.Spec.RuntimeVersion,
									Image: paddle.Spec.StorageURI,
									Ports: []core.ContainerPort{
										{ContainerPort: paddle.Spec.Port, Name: "http", Protocol: "TCP"},
									},
								},
							},
						},
					},
				},
			}
			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "deployment resource should exist")

			Expect(*deployment.Spec.Replicas).To(Equal(int32(1)), "replica count should be equal to 1")

			err = k8sClient.Get(ctx, paddleObjectKey, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to retrieve Paddle resource")

			paddle.Spec.Replicas = pointer.Int32Ptr(2)
			err = k8sClient.Update(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to Update Paddle resource")

			Eventually(getDeploymentReplicasFunc(ctx, deploymentObjectKey)).
				Should(Equal(int32(2)), "expected Deployment resource to be scale to 2 replicas")
		})

		It("should clean up an old Deployment resource if the deploymentName is changed", func() {
			deploymentObjectKey := client.ObjectKey{
				Name:      "deployment-name",
				Namespace: ns.Name,
			}
			newDeploymentObjectKey := client.ObjectKey{
				Name:      "new-deployment",
				Namespace: ns.Name,
			}
			paddleObjectKey := client.ObjectKey{
				Name:      "testresource",
				Namespace: ns.Name,
			}
			paddle := &elasticservingv1.Paddle{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddleObjectKey.Name,
					Namespace: paddleObjectKey.Namespace,
				},
				Spec: elasticservingv1.PaddleSpec{
					DeploymentName: deploymentObjectKey.Name,
					RuntimeVersion: "paddle",
					StorageURI:     "hub.baidubce.com/paddlepaddle/serving:latest",
					Port:           9292,
				},
			}

			err := k8sClient.Create(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to create test Paddle resource")

			deployment := &apps.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      paddle.Spec.DeploymentName,
					Namespace: paddle.Namespace,
				},
				Spec: apps.DeploymentSpec{
					Replicas: paddle.Spec.Replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
						},
					},
					Template: core.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
							},
						},
						Spec: core.PodSpec{
							Containers: []core.Container{
								{
									Name:  paddle.Spec.RuntimeVersion,
									Image: paddle.Spec.StorageURI,
									Ports: []core.ContainerPort{
										{ContainerPort: paddle.Spec.Port, Name: "http", Protocol: "TCP"},
									},
								},
							},
						},
					},
				},
			}
			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "deployment resource should exist")

			err = k8sClient.Get(ctx, paddleObjectKey, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to retrieve Paddle resource")

			paddle.Spec.DeploymentName = newDeploymentObjectKey.Name
			err = k8sClient.Update(ctx, paddle)
			Expect(err).NotTo(HaveOccurred(), "failed to Update Paddle resource")

			Eventually(
				getResourceFunc(ctx, deploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).ShouldNot(BeNil(), "old deployment resource should be deleted")

			Eventually(
				getResourceFunc(ctx, newDeploymentObjectKey, deployment),
				time.Second*5, time.Millisecond*500).Should(BeNil(), "new deployment resource should be created")
		})
	})
})

func getResourceFunc(ctx context.Context, key client.ObjectKey, obj runtime.Object) func() error {
	return func() error {
		return k8sClient.Get(ctx, key, obj)
	}
}

func getDeploymentReplicasFunc(ctx context.Context, key client.ObjectKey) func() int32 {
	return func() int32 {
		depl := &apps.Deployment{}
		err := k8sClient.Get(ctx, key, depl)
		Expect(err).NotTo(HaveOccurred(), "failed to get Deployment resource")

		return *depl.Spec.Replicas
	}
}
