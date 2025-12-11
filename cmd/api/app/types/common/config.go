package common

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func LoadRestConfigFromBearerToken(apiServerURL, bearerToken string) *rest.Config {
	config := &rest.Config{
		Host:        apiServerURL,
		BearerToken: bearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return config
}

func KubeClientSetFromBearerToken(apiServerURL, bearerToken string) (*kubernetes.Clientset, error) {
	config := &rest.Config{
		Host:        apiServerURL,
		BearerToken: bearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	return kubernetes.NewForConfig(config)
}
