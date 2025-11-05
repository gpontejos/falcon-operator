package assets

import (
	falconv1alpha1 "github.com/crowdstrike/falcon-operator/api/falcon/v1alpha1"
	"github.com/crowdstrike/falcon-operator/pkg/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Service returns a Kubernetes Service object
func Service(name string, namespace string, component string, selector map[string]string, portName string, port int32) *corev1.Service {
	labels := common.CRLabels("service", name, component)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       portName,
					Port:       port,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromString(common.FalconServiceHTTPSName),
				},
			},
		},
	}
}

// IARAgentService returns a Kubernetes Service object for the IAR Agent Service
func IARAgentService(falconImageAnalyzer *falconv1alpha1.FalconImageAnalyzer) *corev1.Service {
	labels := common.CRLabels("service", falconImageAnalyzer.Name, common.FalconImageAnalyzer)
	selector := common.CRLabels("deployment", falconImageAnalyzer.Name, common.FalconImageAnalyzer)

	// Set default port if not specified
	httpPort := falconImageAnalyzer.Spec.ImageAnalyzerConfig.IARAgentService.HTTPPort
	if httpPort == 0 {
		httpPort = 8001
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.FalconImageAnalyzerAgentService,
			Namespace: falconImageAnalyzer.Spec.InstallNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       443,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromString("service-port"),
				},
			},
		},
	}
}
