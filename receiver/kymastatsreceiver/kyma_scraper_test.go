package kymastatsreceiver

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

func TestScrape(t *testing.T) {
	gvrs := []schema.GroupVersionResource{
		{
			Group:    "operator.kyma-project.io",
			Version:  "v1",
			Resource: "telemetries",
		},
		{
			Group:    "operator.kyma-project.io",
			Version:  "v1",
			Resource: "istios",
		},
	}

	scheme := runtime.NewScheme()

	telemetry := newUnstructuredObject("operator.kyma-project.io/v1", "Telemetry", "kyma-system", "default")
	unstructured.SetNestedMap(telemetry, map[string]interface{}{
		"state": "Ready",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "TelemetryHealthy",
				"status": "True",
				"reason": "AllFine",
			},
		},
	}, "status")

	istio := newUnstructuredObject("operator.kyma-project.io/v1", "Istio", "kyma-system", "default")
	unstructured.SetNestedMap(istio, map[string]interface{}{
		"state": "Warning",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "IstioHealthy",
				"status": "False",
				"reason": "IstiodDown",
			},
		},
	}, "status")

	// istio custom resource is broken since it has no conditions, thus it should be ignored
	istioCustom := newUnstructuredObject("operator.kyma-project.io/v1", "Istio", "kyma-system", "custom")
	unstructured.SetNestedMap(istioCustom, map[string]interface{}{
		"state": "Ready",
	}, "status")

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			gvrs[0]: "TelemetryList",
			gvrs[1]: "IstioList",
		}, &unstructured.Unstructured{
			Object: telemetry,
		},
		&unstructured.Unstructured{
			Object: istioCustom,
		}, &unstructured.Unstructured{
			Object: istio,
		},
	)

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		gvrs,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	md, err := r.Scrape(context.Background())
	require.NoError(t, err)

	expectedFile := filepath.Join("testdata", "expected_metrics.yaml")
	expected, err := golden.ReadMetrics(expectedFile)

	require.NoError(t, err)
	require.NoError(t, pmetrictest.CompareMetrics(expected, md,
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreStartTimestamp(),
		pmetrictest.IgnoreResourceMetricsOrder(),
		pmetrictest.IgnoreScopeMetricsOrder(),
		pmetrictest.IgnoreMetricsOrder(),
		pmetrictest.IgnoreMetricDataPointsOrder(),
	))
}

func TestScrape_CantPullResource(t *testing.T) {
	gvrs := []schema.GroupVersionResource{
		{
			Group:    "operator.kyma-project.io",
			Version:  "v1",
			Resource: "mykymamodules",
		},
	}

	scheme := runtime.NewScheme()

	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{
		Object: newUnstructuredObject("operator.kyma-project.io/v1", "MyKymaModule", "kyma-system", "default"),
	})

	client.PrependReactor("list", "mykymamodules", func(action clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error")
	})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		gvrs,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.Error(t, err)

}

func TestScrape_HandlesInvalidResourceGracefully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status any
	}{
		{
			name: "no status",
		},
		{
			name:   "status not a map",
			status: "not a map",
		},
		{
			name: "no state",
			status: map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "FakeConditionType",
						"status": "False",
						"reason": "FakeReason",
					},
				},
			},
		},
		{
			name: "state not a string",
			status: map[string]interface{}{
				"state": map[string]interface{}{},
			},
		},
		{
			name: "no conditions",
			status: map[string]interface{}{
				"state": "Ready",
			},
		},
		{
			name: "conditions not a list",
			status: map[string]interface{}{
				"state":      "Ready",
				"conditions": "not a list",
			},
		},
		{
			name: "condition not a map",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					"not a map",
				},
			},
		},
		{
			name: "no condition type",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"status": "False",
						"reason": "FakeReason",
					},
				},
			},
		},
		{
			name: "condition type not a string",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   map[string]interface{}{},
						"status": "False",
						"reason": "FakeReason",
					},
				},
			},
		},
		{
			name: "no condition status",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "FakeConditionType",
						"reason": "FakeReason",
					},
				},
			},
		},
		{
			name: "condition status not a string",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "FakeConditionType",
						"status": map[string]interface{}{},
						"reason": "FakeReason",
					},
				},
			},
		},
		{
			name: "no condition reason",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "FakeConditionType",
						"status": "False",
					},
				},
			},
		},
		{
			name: "condition reason not a string",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "FakeConditionType",
						"status": "False",
						"reason": map[string]interface{}{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gvrs := []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1",
					Resource: "mykymamodules",
				},
			}
			scheme := runtime.NewScheme()
			obj := newUnstructuredObject("operator.kyma-project.io/v1", "MyKymaModule", "kyma-system", "default")
			if tt.status != nil {
				unstructured.SetNestedField(obj, tt.status, "status")
			}

			client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

			r, err := newKymaScraper(
				client,
				receivertest.NewNopSettings(),
				gvrs,
				metadata.DefaultMetricsBuilderConfig(),
			)
			require.NoError(t, err)

			metrics, err := r.Scrape(context.Background())
			require.NoError(t, err)
			require.Zero(t, metrics.DataPointCount())
		})
	}
}

func newUnstructuredObject(apiVersion, kind, namespace, name string) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		},
	}
}
