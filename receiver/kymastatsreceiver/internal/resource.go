package internal

type Resource struct {
	ResourceGroup   string
	ResourceName    string
	ResourceVersion string
}

func NewResource(gorup, name, version string) *Resource {
	return &Resource{
		ResourceGroup:   gorup,
		ResourceName:    name,
		ResourceVersion: version,
	}

}
