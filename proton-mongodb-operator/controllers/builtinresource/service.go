package baseresource

import (
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewService returns a StatefulSet object configured for a name
func NewMongoService(instance *mongodbv1.MongodbOperator) []*corev1.Service {
	ipfamilypolicy := corev1.IPFamilyPolicyRequireDualStack
	var svc = &corev1.Service{}
	svcs := []*corev1.Service{}
	mongodPortName := "mongodb"
	ls := map[string]string{"app": instance.GetName() + "-mongodb"}
	svc = &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", instance.GetName(), "mongodb"),
			Namespace:   instance.Namespace,
			Labels:      map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			Annotations: map[string]string{"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       mongodPortName,
					Port:       28000,
					TargetPort: intstr.IntOrString{IntVal: 28000},
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector:                 ls,
		},
	}
	if instance.Spec.MongoDBSpec.Service.EnableDualStack {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
		svc.Spec.IPFamilyPolicy = &ipfamilypolicy
	}
	svcs = append(svcs, svc)
	if instance.Spec.MongoDBSpec.Service.Type == corev1.ServiceTypeClusterIP {
		svc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-cluster"),
				Namespace: instance.Namespace,
				Labels:    map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       mongodPortName,
						Port:       instance.Spec.MongoDBSpec.Service.Port,
						TargetPort: intstr.IntOrString{IntVal: 28000},
					},
				},
				PublishNotReadyAddresses: true,
				Selector:                 ls,
			},
		}
	} else if instance.Spec.MongoDBSpec.Service.Type == corev1.ServiceTypeNodePort {
		svc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-nodeport"),
				Namespace: instance.Namespace,
				Labels:    map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     mongodPortName,
						Port:     28000,
						NodePort: instance.Spec.MongoDBSpec.Service.Port,
					},
				},
				PublishNotReadyAddresses: true,
				Selector:                 ls,
				Type:                     corev1.ServiceTypeNodePort,
			},
		}
	}
	if instance.Spec.MongoDBSpec.Service.EnableDualStack {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
		svc.Spec.IPFamilyPolicy = &ipfamilypolicy
	}
	svcs = append(svcs, svc)
	return svcs
}

// NewService returns a Service object configured for a name
func NewMgmtService(instance *mongodbv1.MongodbOperator) []*corev1.Service {
	var svc = &corev1.Service{}
	svcs := []*corev1.Service{}
	ipfamilypolicy := corev1.IPFamilyPolicyRequireDualStack
	mgmtPortName := "mgmt"
	ls := map[string]string{"app": instance.GetName() + "-mgmt"}
	svc = &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-mgmt"),
			Namespace:   instance.Namespace,
			Labels:      map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mgmt")},
			Annotations: map[string]string{"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       mgmtPortName,
					Port:       28001,
					TargetPort: intstr.IntOrString{IntVal: 28001},
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector:                 ls,
		},
	}
	if instance.Spec.MgmtSpec.Service.EnableDualStack {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
		svc.Spec.IPFamilyPolicy = &ipfamilypolicy
	}
	svcs = append(svcs, svc)
	if instance.Spec.MgmtSpec.Service.Type == corev1.ServiceTypeClusterIP {
		svc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mgmt-cluster"),
				Namespace: instance.Namespace,
				Labels:    map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       mgmtPortName,
						Port:       instance.Spec.MgmtSpec.Service.Port,
						TargetPort: intstr.IntOrString{IntVal: 28001},
					},
				},
				Selector: ls,
			},
		}
	} else if instance.Spec.MgmtSpec.Service.Type == corev1.ServiceTypeNodePort {
		svc = &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mgmt-nodeport"),
				Namespace: instance.Namespace,
				Labels:    map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:     mgmtPortName,
						Port:     28001,
						NodePort: instance.Spec.MgmtSpec.Service.Port,
					},
				},
				Type:     corev1.ServiceTypeNodePort,
				Selector: ls,
			},
		}
	}
	if instance.Spec.MgmtSpec.Service.EnableDualStack {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
		svc.Spec.IPFamilyPolicy = &ipfamilypolicy
	}
	svcs = append(svcs, svc)
	return svcs
}

// NewExporterService returns a ExporterService object configured for a name
func NewExporterService(instance *mongodbv1.MongodbOperator) *corev1.Service {
	ipfamilypolicy := corev1.IPFamilyPolicyRequireDualStack
	exporterPortName := "exporter"
	ls := map[string]string{"app": instance.GetName() + "-mongodb-exporter"}
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", instance.GetName(), "mongodb-exporter"),
			Namespace: instance.Namespace,
			Labels:    map[string]string{"app": fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")},
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "9216",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       exporterPortName,
					Port:       9216,
					TargetPort: intstr.IntOrString{IntVal: 9216},
				},
			},
			Selector:  ls,
			ClusterIP: "None",
		},
	}
	if instance.Spec.MongoDBSpec.Service.EnableDualStack {
		svc.Spec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol}
		svc.Spec.IPFamilyPolicy = &ipfamilypolicy
	}
	return svc
}
