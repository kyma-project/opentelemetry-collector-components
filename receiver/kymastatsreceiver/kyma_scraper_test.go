package kymastatsreceiver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

const (
	dataLen = 6
)

var (
	statusError = map[string]interface{}{
		"state": "Error",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "ConditionType1",
				"status":  "False",
				"reason":  "Reason1",
				"message": "some message",
			},
		},
	}

	statusReady = map[string]interface{}{
		"state": "Ready",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "ConditionType1",
				"status":  "False",
				"reason":  "Reason1",
				"message": "some message",
			},
		},
	}
)

func TestScrape(t *testing.T) {
	rcConfig := []schema.GroupVersionResource{
		{
			Group:    "group",
			Version:  "version",
			Resource: "thekinds",
		},
	}

	scheme := runtime.NewScheme()

	obj1 := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")
	obj2 := newUnstructuredObject("group2/version", "TheKind", "ns-foo", "name2-foo")
	obj3 := newUnstructuredObject("group/version", "TheKind", "ns-foo1", "name-bar")
	obj4 := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-baz")
	obj5 := newUnstructuredObject("group2/version", "TheKind", "ns-foo", "name2-baz")
	obj6 := newUnstructuredObject("group/version", "AnotherKind", "ns-foo", "name2-baz")

	unstructured.SetNestedMap(obj1, statusError, "status")
	unstructured.SetNestedMap(obj2, statusError, "status")
	unstructured.SetNestedMap(obj3, statusReady, "status")
	unstructured.SetNestedMap(obj4, statusReady, "status")
	unstructured.SetNestedMap(obj5, statusReady, "status")
	unstructured.SetNestedMap(obj6, statusReady, "status")

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "group", Version: "version", Resource: "thekinds"}: "TheKindList",
		}, &unstructured.Unstructured{
			Object: obj1,
		},
		&unstructured.Unstructured{
			Object: obj2,
		}, &unstructured.Unstructured{
			Object: obj3,
		}, &unstructured.Unstructured{
			Object: obj4,
		}, &unstructured.Unstructured{
			Object: obj5,
		}, &unstructured.Unstructured{
			Object: obj6},
	)

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	md, err := r.Scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, dataLen, md.DataPointCount())
}

func TestScrape_CantPullResource(t *testing.T) {
	rcConfig := []schema.GroupVersionResource{
		{
			Group:    "group",
			Version:  "version",
			Resource: "thekinds",
		},
	}

	scheme := runtime.NewScheme()

	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{
		Object: newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo"),
	})

	client.PrependReactor("list", "thekinds", func(action clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error")
	})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
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

			rcConfig := []schema.GroupVersionResource{
				{
					Group:    "operator.kyma-project.io",
					Version:  "v1",
					Resource: "mykymamodules",
				},
			}
			scheme := runtime.NewScheme()
			obj := newUnstructuredObject("operator.kyma-project.io/v1", "MyKymaModule", "default", "default")
			if tt.status != nil {
				unstructured.SetNestedField(obj, tt.status, "status")
			}

			client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

			r, err := newKymaScraper(
				client,
				receivertest.NewNopSettings(),
				rcConfig,
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
