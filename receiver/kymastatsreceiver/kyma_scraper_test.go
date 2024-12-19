package kymastatsreceiver

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

const (
	telemetryResourceGroup     = "operator.kyma-project.io"
	telemetryResourceVersion   = "v1"
	telemetryResourceNamespace = "kyma-system"

	logPipelineResourceGroup   = "telemetry.kyma-project.io"
	logPipelineResourceVersion = "v1alpha1"
)

func TestScrape(t *testing.T) {
	resources := []ResourceConfig{
		{
			Group:    telemetryResourceGroup,
			Version:  telemetryResourceVersion,
			Resource: "telemetries",
		},
		{
			Group:    logPipelineResourceGroup,
			Version:  logPipelineResourceVersion,
			Resource: "logpipelines",
		},
	}

	scheme := runtime.NewScheme()

	telemetry := newUnstructuredObject("Telemetry", "telemetry", "default")
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

	pipe1 := newUnstructuredObject("LogPipeline", "logpipeline", "pipe-1")
	pipe2 := newUnstructuredObject("LogPipeline", "logpipeline", "pipe-2")

	unstructured.SetNestedMap(pipe1, map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "AgentHealthy",
				"status": "True",
				"reason": "AgentReady",
			},
		},
	}, "status")

	unstructured.SetNestedMap(pipe2, map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "AgentHealthy",
				"status": "False",
				"reason": "AgentNotReady",
			},
		},
	}, "status")

	dynamic := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			schema.GroupVersionResource(resources[0]): "TelemetryList",
			schema.GroupVersionResource(resources[1]): "LogPipelineList",
		}, &unstructured.Unstructured{
			Object: telemetry,
		},
		&unstructured.Unstructured{
			Object: pipe1,
		},
		&unstructured.Unstructured{
			Object: pipe2,
		},
	)

	r, err := newKymaScraper(
		Config{
			MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
			Resources:            resources,
		},
		dynamic,
		receivertest.NewNopSettings(),
	)

	require.NoError(t, err)

	md, err := r.ScrapeMetrics(context.Background())
	require.NoError(t, err)

	expectedFile := filepath.Join("testdata", "metrics.yaml")
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
	resources := []ResourceConfig{
		{
			Group:    telemetryResourceGroup,
			Version:  telemetryResourceVersion,
			Resource: "mykymaresources",
		},
	}

	scheme := runtime.NewScheme()

	dynamic := dynamicfake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{
		Object: newUnstructuredObject("MyKymaResource", "telemetry", "default"),
	})

	dynamic.PrependReactor("list", "mykymaresources", func(action clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error")
	})

	r, err := newKymaScraper(
		Config{
			MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
			Resources:            resources,
		},
		dynamic,
		receivertest.NewNopSettings(),
	)

	require.NoError(t, err)

	_, err = r.ScrapeMetrics(context.Background())
	require.Error(t, err)

}

func TestScrape_HandlesInvalidResourceGracefully(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		status             any
		expectedDataPoints int
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
		},
		{
			name: "conditions not a list",
			status: map[string]interface{}{
				"state":      "Ready",
				"conditions": "not a list",
			},
			expectedDataPoints: 1,
		},
		{
			name: "condition not a map",
			status: map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					"not a map",
				},
			},
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
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
			expectedDataPoints: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resources := []ResourceConfig{
				{
					Group:    telemetryResourceGroup,
					Version:  telemetryResourceVersion,
					Resource: "mykymaresources",
				},
			}
			scheme := runtime.NewScheme()
			obj := newUnstructuredObject("MyKymaResource", "telemetry", "default")
			if tt.status != nil {
				unstructured.SetNestedField(obj, tt.status, "status")
			}

			dynamic := dynamicfake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

			r, err := newKymaScraper(
				Config{
					MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
					Resources:            resources,
				},
				dynamic,
				receivertest.NewNopSettings(),
			)
			require.NoError(t, err)

			metrics, err := r.ScrapeMetrics(context.Background())
			require.NoError(t, err)
			require.Equal(t, tt.expectedDataPoints, metrics.DataPointCount())
		})
	}
}

func newUnstructuredObject(kind, resourceType, name string) map[string]interface{} {
	if resourceType == "telemetry" {
		return map[string]interface{}{
			"apiVersion": telemetryResourceGroup + "/" + telemetryResourceVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": telemetryResourceNamespace,
				"name":      name,
			},
		}
	}
	if resourceType == "logpipeline" {
		return map[string]interface{}{
			"apiVersion": logPipelineResourceGroup + "/" + logPipelineResourceVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name,
			},
		}
	}
	return nil
}
