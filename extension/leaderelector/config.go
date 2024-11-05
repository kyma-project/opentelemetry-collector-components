package leaderelector

import (
	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"k8s.io/client-go/kubernetes"
	"time"
)

// Config is the configuration struct for your extension.
type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`
	leaseName           string        `mapstructure:"lease_name"`
	leaseNamespace      string        `mapstructure:"lease_namespace"`
	leaseDuration       time.Duration `mapstructure:"lease_duration"`
	renewDuration       time.Duration `mapstructure:"renew_deadline"`
	retryPeriod         time.Duration `mapstructure:"retry_period"`

	makeClient func() (kubernetes.Interface, error)
	// Define any custom fields for your extension's configuration
}

func (cfg *Config) getK8sClient() (kubernetes.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeClient()
	}
	return k8sconfig.MakeClient(cfg.APIConfig)
}
