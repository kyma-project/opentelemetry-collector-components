package k8sconfig

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AuthType describes the type of authentication to use for the K8s API
type AuthType string

const (
	AuthTypeServiceAccount AuthType = "serviceAccount"
)

func GetK8sClient(authType AuthType) (kubernetes.Interface, error) {
	if authType != AuthTypeServiceAccount {
		return nil, fmt.Errorf("authentication type: %s not supported", string(authType))
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
