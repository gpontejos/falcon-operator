package common

import (
	"github.com/crowdstrike/falcon-operator/api/falcon/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type FalconCRD interface {
	*v1alpha1.FalconNodeSensor | *v1alpha1.FalconContainer | *v1alpha1.FalconAdmission | *v1alpha1.FalconImageAnalyzer

	metav1.Object
	runtime.Object
	GetFalconSecretSpec() v1alpha1.FalconSecret
	GetFalconAPISpec() *v1alpha1.FalconAPI
	SetFalconAPISpec(*v1alpha1.FalconAPI)
	GetFalconSpec() v1alpha1.FalconSensor
	SetFalconSpec(v1alpha1.FalconSensor)
	Tolerations() *[]corev1.Toleration
}
