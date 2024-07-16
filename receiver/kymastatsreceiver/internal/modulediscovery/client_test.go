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

func TestDiscover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resources []*metav1.APIResourceList
		expected  []schema.GroupVersionResource
	}{
		{
			name: "preffered version without subresource",
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
				{
					GroupVersion: "operator.kyma-project.io/v1alpha1",
					APIResources: []metav1.APIResource{
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
			name: "resource only available in old group version",
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
					Version:  "v1apha1",
					Resource: "telemetries",
				}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			discovery := discoveryfake.FakeDiscovery{
				Fake: &clienttesting.Fake{
					Resources: test.resources,
				},
			}
			sut := New(&discovery, zap.NewNop(), []string{"operator.kyma-project.io"})

			gvrs, err := sut.Discover()
			require.NoError(t, err)

			require.ElementsMatch(t, test.expected, gvrs)
		})
	}
}
