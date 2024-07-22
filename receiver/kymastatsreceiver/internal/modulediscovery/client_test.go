package modulediscovery

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discoveryfake "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
)

type fakeDiscovery struct {
	discoveryfake.FakeDiscovery
}

// ServerPreferredResources returns predefined resources (FakeDiscovery returns hard-coded nil,nil for some reason)
func (fd *fakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return fd.Resources, nil
}

func TestDiscover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		resources         []*metav1.APIResourceList
		excludedResources []string
		expected          []schema.GroupVersionResource
	}{
		{
			name: "without subresource",
			resources: []*metav1.APIResourceList{
				{
					GroupVersion: "operator.kyma-project.io/v1beta1",
					APIResources: []metav1.APIResource{
						{Name: "istios"},
						{Name: "istios/scale"},
						{Name: "telemetries"},
						{Name: "telemetries/scale"},
					},
				},
			},
			expected: []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1beta1",
					Resource: "istios",
				},
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1beta1",
					Resource: "telemetries",
				}},
		},
		{
			name: "multiple group versions",
			resources: []*metav1.APIResourceList{
				{
					GroupVersion: "operator.kyma-project.io/v1beta1",
					APIResources: []metav1.APIResource{
						{Name: "istios"},
					},
				},
				{
					GroupVersion: "operator.kyma-project.io/v1alpha1",
					APIResources: []metav1.APIResource{
						{Name: "telemetries"},
					},
				},
			},
			expected: []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1beta1",
					Resource: "istios",
				},
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1alpha1",
					Resource: "telemetries",
				}},
		},

		{
			name: "unknown group versions",
			resources: []*metav1.APIResourceList{
				{
					GroupVersion: "operator.kyma-project.io/v1beta1",
					APIResources: []metav1.APIResource{
						{Name: "istios"},
					},
				},
				{
					GroupVersion: "telemetry.istio.io/v1alpha1",
					APIResources: []metav1.APIResource{
						{Name: "telemetries"},
					},
				},
			},
			expected: []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1beta1",
					Resource: "istios",
				},
			},
		},
		{
			name: "excluded resources",
			resources: []*metav1.APIResourceList{
				{
					GroupVersion: "operator.kyma-project.io/v1beta1",
					APIResources: []metav1.APIResource{
						{Name: "istios"},
						{Name: "kymas"},
					},
				},
			},
			excludedResources: []string{"kymas"},
			expected: []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1beta1",
					Resource: "istios",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			discovery := fakeDiscovery{
				FakeDiscovery: discoveryfake.FakeDiscovery{
					Fake: &clienttesting.Fake{
						Resources: test.resources,
					},
				},
			}
			sut := New(&discovery, zap.NewNop(), Config{
				ExludedResources: test.excludedResources,
				ModuleGroups:     []string{"operator.kyma-project.io"},
			})

			gvrs, err := sut.Discover()
			require.NoError(t, err)

			require.ElementsMatch(t, test.expected, gvrs)
		})
	}
}
