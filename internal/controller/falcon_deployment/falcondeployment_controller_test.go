package falcon

import (
	"context"
	"reflect"
	"time"

	falconv1alpha1 "github.com/crowdstrike/falcon-operator/api/falcon/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	FalconDeploymentName              = "test-falcondeployment"
	falconClientID                    = "12345678912345678912345678912345"
	falconSecret                      = "32165498732165498732165498732165498732165"
	falconCID                         = "1234567890ABCDEF1234567890ABCDEF-12"
	falconCloudRegion                 = "us-1"
	defaultRegistryName               = "crowdstrike"
	overrideRegistryName              = "ecr"
	overrideFalconClientID            = "11111111111111111111111111111111"
	overrideFalconSecret              = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	overrideFalconCID                 = "11111111111111111111111111111111-FF"
	overrideFalconCloudRegion         = "us-2"
	sidecarSensorNamespace            = "falcon-system"
	sidecarSensorNamespacedName       = "falcon-container-sensor"
	nodeSensorNamespace               = "falcon-node-sensor"
	nodeSensorNamespacedName          = "falcon-node-sensor"
	imageAnalyzerNamespace            = "falcon-iar"
	imageAnalyzerNamespacedName       = "falcon-image-analyzer"
	admissionControllerNamespace      = "falcon-kac"
	admissionControllerNamespacedName = "falcon-kac"
)

var _ = Describe("FalconDeployment Controller", func() {
	Context("FalconDeployment controller test", func() {
		defaultRegistry := falconv1alpha1.RegistrySpec{
			Type: falconv1alpha1.RegistryTypeSpec(defaultRegistryName),
		}

		overrideRegistry := falconv1alpha1.RegistrySpec{
			Type: falconv1alpha1.RegistryTypeSpec(overrideRegistryName),
		}

		mockFalconAPI := falconv1alpha1.FalconAPI{
			CloudRegion:  falconCloudRegion,
			ClientId:     falconClientID,
			ClientSecret: falconSecret,
			CID:          &falconCID,
		}

		overrideFalconAPI := falconv1alpha1.FalconAPI{
			CloudRegion:  overrideFalconCloudRegion,
			ClientId:     overrideFalconClientID,
			ClientSecret: overrideFalconSecret,
			CID:          &overrideFalconCID,
		}

		ctx := context.Background()

		falconDeploymentNamespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      FalconDeploymentName,
				Namespace: FalconDeploymentName,
			},
		}

		typeContainerNamespacedName := types.NamespacedName{Name: sidecarSensorNamespacedName, Namespace: sidecarSensorNamespace}
		typeAdmissionNamespacedName := types.NamespacedName{Name: admissionControllerNamespacedName, Namespace: admissionControllerNamespace}
		typeIARNamespacedName := types.NamespacedName{Name: imageAnalyzerNamespacedName, Namespace: imageAnalyzerNamespace}
		typeNodeNamespacedName := types.NamespacedName{Name: nodeSensorNamespacedName, Namespace: nodeSensorNamespace}
		typeNamespacedName := types.NamespacedName{Name: FalconDeploymentName, Namespace: FalconDeploymentName}

		BeforeAll(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, falconDeploymentNamespace)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterAll(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, falconDeploymentNamespace)
		})

		It("should successfully reconcile the resource", func() {
			By("Creating the custom resource for the Kind FalconDeployment - No Overrides")
			deployContainerSensor := true

			falconDeployment := &falconv1alpha1.FalconDeployment{}
			err := k8sClient.Get(ctx, typeNamespacedName, falconDeployment)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				falconDeployment := &falconv1alpha1.FalconDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      FalconDeploymentName,
						Namespace: falconDeploymentNamespace.Name,
					},
					Spec: falconv1alpha1.FalconDeploymentSpec{
						FalconAPI:             &mockFalconAPI,
						Registry:              defaultRegistry,
						DeployContainerSensor: &deployContainerSensor,
						FalconAdmission: falconv1alpha1.FalconAdmissionSpec{
							InstallNamespace: admissionControllerNamespace,
							Registry:         defaultRegistry,
						},
						FalconNodeSensor: falconv1alpha1.FalconNodeSensorSpec{
							InstallNamespace: nodeSensorNamespace,
						},
						FalconImageAnalyzer: falconv1alpha1.FalconImageAnalyzerSpec{
							InstallNamespace: imageAnalyzerNamespace,
							Registry:         defaultRegistry,
						},
						FalconContainerSensor: falconv1alpha1.FalconContainerSpec{
							InstallNamespace: sidecarSensorNamespace,
							Registry:         defaultRegistry,
						},
					},
				}

				err = k8sClient.Create(ctx, falconDeployment)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if FalconDeployment was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconDeployment{}
				return k8sClient.Get(ctx, typeNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the FalconDeployment custom resource created")
			falconDeploymentReconciler := &FalconDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = falconDeploymentReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if FalconNodeSensor was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconNodeSensor{}
				return k8sClient.Get(ctx, typeNodeNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconAdmission was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconAdmission{}
				return k8sClient.Get(ctx, typeAdmissionNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconImageAnalyzer was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconImageAnalyzer{}
				return k8sClient.Get(ctx, typeIARNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconContainer was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconContainer{}
				return k8sClient.Get(ctx, typeContainerNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Validate default FalconAPI credentials are copied to child CRs - FalconAdmission")
			falconAdmission := &falconv1alpha1.FalconAdmission{}
			Expect(reflect.DeepEqual(falconDeployment.Spec.FalconAPI, falconAdmission.Spec.FalconAPI)).To(Equal(true))

			By("Validate default FalconAPI credentials are copied to child CRs - FalconImageAnalyzer")
			falconImageAnalyzer := &falconv1alpha1.FalconImageAnalyzer{}
			Expect(reflect.DeepEqual(falconDeployment.Spec.FalconAPI, falconImageAnalyzer.Spec.FalconAPI)).To(Equal(true))

			By("Validate default FalconAPI credentials are copied to child CRs - FalconNodeSensor")
			falconNodeSensor := &falconv1alpha1.FalconNodeSensor{}
			Expect(reflect.DeepEqual(falconDeployment.Spec.FalconAPI, falconNodeSensor.Spec.FalconAPI)).To(Equal(true))

			By("Validate default FalconAPI credentials are copied to child CRs - FalconContainer")
			falconContainer := &falconv1alpha1.FalconContainer{}
			Expect(reflect.DeepEqual(falconDeployment.Spec.FalconAPI, falconContainer.Spec.FalconAPI)).To(Equal(true))

			By("Deleting the custom resource for the Kind FalconDeployment - No Overrides")
		})

		It("should successfully reconcile the resource", func() {
			By("Creating the custom resource for the Kind FalconDeployment - With Overrides")
			deployContainerSensor := true

			falconDeployment := &falconv1alpha1.FalconDeployment{}
			err := k8sClient.Get(ctx, typeNamespacedName, falconDeployment)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				falconDeployment := &falconv1alpha1.FalconDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      FalconDeploymentName,
						Namespace: falconDeploymentNamespace.Name,
					},
					Spec: falconv1alpha1.FalconDeploymentSpec{
						FalconAPI:             &mockFalconAPI,
						Registry:              defaultRegistry,
						DeployContainerSensor: &deployContainerSensor,
						FalconAdmission: falconv1alpha1.FalconAdmissionSpec{
							InstallNamespace: admissionControllerNamespace,
							FalconAPI:        &overrideFalconAPI,
							Registry:         overrideRegistry,
						},
						FalconNodeSensor: falconv1alpha1.FalconNodeSensorSpec{
							InstallNamespace: nodeSensorNamespace,
							FalconAPI:        &overrideFalconAPI,
						},
						FalconImageAnalyzer: falconv1alpha1.FalconImageAnalyzerSpec{
							InstallNamespace: imageAnalyzerNamespace,
							FalconAPI:        &overrideFalconAPI,
							Registry:         overrideRegistry,
						},
						FalconContainerSensor: falconv1alpha1.FalconContainerSpec{
							InstallNamespace: sidecarSensorNamespace,
							FalconAPI:        &overrideFalconAPI,
							Registry:         overrideRegistry,
						},
					},
				}

				err = k8sClient.Create(ctx, falconDeployment)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if FalconDeployment was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconDeployment{}
				return k8sClient.Get(ctx, typeNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the FalconDeployment custom resource created")
			falconDeploymentReconciler := &FalconDeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = falconDeploymentReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if FalconNodeSensor was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconNodeSensor{}
				return k8sClient.Get(ctx, typeNodeNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconAdmission was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconAdmission{}
				return k8sClient.Get(ctx, typeAdmissionNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconImageAnalyzer was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconImageAnalyzer{}
				return k8sClient.Get(ctx, typeIARNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if FalconContainer was created")
			Eventually(func() error {
				found := &falconv1alpha1.FalconContainer{}
				return k8sClient.Get(ctx, typeContainerNamespacedName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Validate override FalconAPI credentials are used in the child CRs - FalconAdmission")
			falconAdmission := &falconv1alpha1.FalconAdmission{}
			Expect(reflect.DeepEqual(falconAdmission.Spec.FalconAPI, overrideFalconAPI)).To(Equal(true))

			By("Validate override FalconAPI credentials are used in the child CRs - FalconImageAnalyzer")
			falconImageAnalyzer := &falconv1alpha1.FalconImageAnalyzer{}
			Expect(reflect.DeepEqual(falconImageAnalyzer.Spec.FalconAPI, overrideFalconAPI)).To(Equal(true))

			By("Validate override FalconAPI credentials are used in the child CRss - FalconNodeSensor")
			falconNodeSensor := &falconv1alpha1.FalconNodeSensor{}
			Expect(reflect.DeepEqual(falconNodeSensor.Spec.FalconAPI, overrideFalconAPI)).To(Equal(true))

			By("Validate override FalconAPI credentials are used in the child CRs - FalconContainer")
			falconContainer := &falconv1alpha1.FalconContainer{}
			Expect(reflect.DeepEqual(falconContainer.Spec.FalconAPI, overrideFalconAPI)).To(Equal(true))
		})
	})
})
