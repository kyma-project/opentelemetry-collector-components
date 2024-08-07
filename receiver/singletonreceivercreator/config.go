package singletonreceivercreator

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

const (
	// receiversConfigKey is the config key name used to specify the subreceivers.
	subreceiverConfigKey = "receiver"

	leaderElectionConfigKey = "leader_election"
)

var (
	errMissingLeaseName      = errors.New(`"leaseName" not specified in config`)
	errMissingLeaseNamespace = errors.New(`"leaseNamespace" not specified in config`)
	errNonPositiveInterval   = errors.New("requires positive value")
)

// Config defines configuration for leader receiver creator.
type Config struct {
	k8sconfig.APIConfig  `mapstructure:",squash"`
	leaderElectionConfig leaderElectionConfig `yaml:"leader_election"`
	subreceiverConfig    receiverConfig

	makeClient func() (kubernetes.Interface, error)
}

type leaderElectionConfig struct {
	leaseName      string        `mapstructure:"lease_name"`
	leaseNamespace string        `mapstructure:"lease_namespace"`
	leaseDuration  time.Duration `mapstructure:"lease_duration"`
	renewDuration  time.Duration `mapstructure:"renew_deadline"`
	retryPeriod    time.Duration `mapstructure:"retry_period"`
}

// receiverConfig describes a receiver instance with a default config.
type receiverConfig struct {
	// id is the id of the subreceiver (ie <receiver type>/<id>).
	id component.ID
	// config is the map configured by the user in the config file. It is the contents of the map from
	// the "config" section. The keys and values are arbitrarily configured by the user.
	config map[string]any
}

var _ confmap.Unmarshaler = (*Config)(nil)

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		// Nothing to do if there is no config given.
		return nil
	}

	if err := componentParser.Unmarshal(cfg, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	subreceiverConfig, err := componentParser.Sub(subreceiverConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", subreceiverConfigKey, err)
	}

	lec, err := componentParser.Sub(leaderElectionConfigKey)
	if err != nil {
		return fmt.Errorf("unable to extract key %v: %w", leaderElectionConfigKey, err)
	}

	cfg.leaderElectionConfig, err = unmarshalLeaderElectionConfig(cfg.leaderElectionConfig, lec.ToStringMap())
	if err != nil {
		return fmt.Errorf("failed to create leader election config: %w", err)
	}

	for subreceiverKey := range subreceiverConfig.ToStringMap() {
		receiverConfig, err := subreceiverConfig.Sub(subreceiverKey)
		if err != nil {
			return fmt.Errorf("unable to extract subreceiver key %v: %w", subreceiverKey, err)
		}

		cfg.subreceiverConfig, err = newReceiverConfig(subreceiverKey, receiverConfig.ToStringMap())
		if err != nil {
			return fmt.Errorf("failed to create subreceiver config: %w", err)
		}

		return nil
	}

	return nil
}

// and its arbitrary config map values.
func newReceiverConfig(name string, cfg map[string]any) (receiverConfig, error) {
	id := component.ID{}
	if err := id.UnmarshalText([]byte(name)); err != nil {
		return receiverConfig{}, fmt.Errorf("failed to parse subreceiver id %v: %w", name, err)
	}

	return receiverConfig{
		id:     id,
		config: cfg,
	}, nil
}

func unmarshalLeaderElectionConfig(lec leaderElectionConfig, cfg map[string]any) (leaderElectionConfig, error) {
	if leaseName, ok := cfg["lease_name"].(string); ok {
		lec.leaseName = leaseName
	}

	if leaseNamespace, ok := cfg["lease_namespace"].(string); ok {
		lec.leaseNamespace = leaseNamespace
	}

	if leaseDuration, ok := cfg["lease_duration"].(string); ok {
		leaseDuration, err := time.ParseDuration(leaseDuration)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse lease duration: %w", err)
		}
		lec.leaseDuration = leaseDuration
	}

	if renewDeadline, ok := cfg["renew_deadline"].(string); ok {
		renewDeadline, err := time.ParseDuration(renewDeadline)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse renew deadline: %w", err)
		}
		lec.renewDuration = renewDeadline
	}

	if retryPeriod, ok := cfg["retry_period"].(string); ok {
		retryPeriod, err := time.ParseDuration(retryPeriod)
		if err != nil {
			return leaderElectionConfig{}, fmt.Errorf("failed to parse retry period: %w", err)
		}
		lec.retryPeriod = retryPeriod
	}

	return lec, nil
}

func (cfg *Config) Validate() error {
	if err := cfg.leaderElectionConfig.validate(); err != nil {
		return err
	}

	if err := cfg.APIConfig.Validate(); err != nil {
		return err
	}

	return nil
}

func (lec *leaderElectionConfig) validate() error {
	if lec.leaseName == "" {
		return errMissingLeaseName
	}

	if lec.leaseNamespace == "" {
		return errMissingLeaseNamespace
	}

	if lec.leaseDuration <= 0 {
		return fmt.Errorf(`"lease_duration": %w`, errNonPositiveInterval)
	}

	if lec.renewDuration <= 0 {
		return fmt.Errorf(`"renew_deadline": %w`, errNonPositiveInterval)
	}

	if lec.retryPeriod <= 0 {
		return fmt.Errorf(`"retry_period": %w`, errNonPositiveInterval)
	}

	return nil
}

func (cfg *Config) getK8sClient() (kubernetes.Interface, error) {
	if cfg.makeClient != nil {
		return cfg.makeClient()
	}
	return k8sconfig.MakeClient(cfg.APIConfig)
}
