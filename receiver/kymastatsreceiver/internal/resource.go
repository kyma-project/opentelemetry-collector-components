package internal

type Resource struct {
	ResourceGroup   string `mapstructure:"resource_group"`
	ResourceName    string `mapstructure:"resource_name"`
	ResourceVersion string `mapstructure:"resource_version"`
}

func NewDefaultResourceConfiguration() []Resource {
	return []Resource{
		{
			ResourceGroup:   "operator.kyma-project.io",
			ResourceName:    "Telemetry",
			ResourceVersion: "v1alpha1",
		},
	}
}
