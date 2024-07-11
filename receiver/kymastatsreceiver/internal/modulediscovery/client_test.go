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
	discovery := discoveryfake.FakeDiscovery{
		Fake: &clienttesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: "operator.kyma-project.io/v1beta1",
					APIResources: []metav1.APIResource{
						{Name: "istio"},
						{Name: "istio/scale"},
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
		},
	}
	sut := New(&discovery, zap.NewNop(), []string{"operator.kyma-project.io"})

	gvrs, err := sut.Discover()
	require.NoError(t, err)

	require.Len(t, gvrs, 2, "expect gvrs of preferred version without subresources")
	require.Equal(t, schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta1",
		Resource: "istio",
	}, gvrs[0])
	require.Equal(t, schema.GroupVersionResource{
		Group:    "operator.kyma-project.io",
		Version:  "v1beta1",
		Resource: "telemetries",
	}, gvrs[1])
}
