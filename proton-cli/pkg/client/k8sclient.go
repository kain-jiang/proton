package client

import (
	"errors"
	"net/http"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ErrKubernetesDynamicClientNil   = errors.New("kubernetes dynamic client is nil")
	ErrKubernetesClientSetNil       = errors.New("kubernetes client set is nil")
	ErrKubernetesExtensionClientNil = errors.New("kubernetes extension client is nil")
)

func NewK8sClientInterface() (dynamic.Interface, kubernetes.Interface) {
	clientDynamic, clientSet := NewK8sClient()
	if clientSet == nil {
		return clientDynamic, nil
	}
	return clientDynamic, clientSet
}

func NewK8sHTTPClient() (*http.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}
	return rest.HTTPClientFor(config)
}

func NewK8sClient() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return
	}
	clientDynamic, _ = dynamic.NewForConfig(config)
	clientSet, _ = kubernetes.NewForConfig(config)
	return
}

func NewExtK8sClient() (client *apiextensionsv1.ApiextensionsV1Client) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return
	}
	client, _ = apiextensionsv1.NewForConfig(config)
	return
}

const ResourceRDSMariaDBCluster = "rdsmariadbclusters"
const ResourceMongoDBCluster = "mongodboperators"
