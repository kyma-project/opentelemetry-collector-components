package k8sconfig

import (
	"fmt"
	"net/http"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// AuthType describes the type of authentication to use for the K8s API
type AuthType string

const (
	AuthTypeServiceAccount AuthType = "serviceAccount"
	// AuthTypeKubeConfig uses local credentials like those used by kubectl.
	AuthTypeKubeConfig AuthType = "kubeConfig"
)

var authTypes = map[AuthType]bool{
	AuthTypeServiceAccount: true,
	AuthTypeKubeConfig:     true,
}

// APIConfig contains options relevant to connecting to the K8s API
type APIConfig struct {
	// How to authenticate to the K8s API server.  This can be one of
	// `serviceAccount` (to use the standard service account
	// token provided to the agent pod), or `kubeConfig` to use credentials
	// from `~/.kube/config`.
	AuthType AuthType `mapstructure:"auth_type"`

	// When using auth_type `kubeConfig`, override the current context.
	Context string `mapstructure:"context"`
}

// Validate validates the K8s API config
func (c APIConfig) Validate() error {
	if !authTypes[c.AuthType] {
		return fmt.Errorf("invalid authType for kubernetes: %v", c.AuthType)
	}

	return nil
}

// CreateRestConfig creates an Kubernetes API config from user configuration.
func CreateRestConfig(apiConf APIConfig) (*rest.Config, error) {
	var authConf *rest.Config
	var err error

	authType := apiConf.AuthType

	switch authType {
	case AuthTypeKubeConfig:
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		if apiConf.Context != "" {
			configOverrides.CurrentContext = apiConf.Context
		}
		authConf, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules, configOverrides).ClientConfig()

		if err != nil {
			return nil, fmt.Errorf("error connecting to k8s with auth_type=%s: %w", AuthTypeKubeConfig, err)
		}

	case AuthTypeServiceAccount:
		// This should work for most clusters but other auth types can be added
		authConf, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("authentication type not supported")

	}

	authConf.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		// Don't use system proxy settings since the API is local to the
		// cluster
		if t, ok := rt.(*http.Transport); ok {
			t.Proxy = nil
		}
		return rt
	}

	return authConf, nil
}

// MakeClient can take configuration if needed for other types of auth
func MakeClient(apiConf APIConfig) (k8s.Interface, error) {
	if err := apiConf.Validate(); err != nil {
		return nil, err
	}

	authConf, err := CreateRestConfig(apiConf)
	if err != nil {
		return nil, err
	}

	client, err := k8s.NewForConfig(authConf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
