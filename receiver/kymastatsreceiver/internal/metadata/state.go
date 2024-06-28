package metadata

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ResourceStatusData struct {
	State      string
	Conditions []v1.Condition
	Name       string
	Namespace  string
}

type Stats struct {
	Resources []ResourceStatusData
}
