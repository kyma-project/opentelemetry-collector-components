package metadata

type ResourceStatusData struct {
	State      string
	Conditions []Condition
	Name       string
	Namespace  string
}

type Condition struct {
	Type   string
	Status string
	Reason string
}

type Stats struct {
	Resources []ResourceStatusData
}
