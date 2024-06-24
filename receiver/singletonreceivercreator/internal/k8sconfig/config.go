package k8sconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// This package has been copied from "https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/internal/k8sconfig/config.go"
// some modifications have been made to the original code to better suit the needs of this project. Additionally, importing internal packages
// from other modules is not supported in golang.

// AuthType describes the type of authentication to use for the K8s API
type AuthType string

const (
	// AuthTypeServiceAccount means to use the built-in service account that
	// K8s automatically provisions for each pod.
	AuthTypeServiceAccount AuthType = "serviceAccount"
	// AuthTypeKubeConfig uses local credentials like those used by kubectl.
	AuthTypeKubeConfig AuthType = "kubeConfig"
)

var authTypes = map[AuthType]bool{
	AuthTypeServiceAccount: true,
	AuthTypeKubeConfig:     true,
}

func GetK8sClient(authType AuthType) (kubernetes.Interface, error) {
	if !authTypes[authType] {
		return nil, fmt.Errorf("authentication type: %s not supported", string(authType))
	}
	var config *rest.Config
	var err error
	switch authType {
	case AuthTypeServiceAccount:
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	case AuthTypeKubeConfig:
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			kubeConfigPath = filepath.Join(os.Getenv("HOME"), ".kube/config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
