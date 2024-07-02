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

const dataLen = 6

func TestScraper(t *testing.T) {
	rcConfig := []Resource{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}

	scheme := runtime.NewScheme()

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			{Group: "group", Version: "version", Resource: "thekinds"}: "TheKindList",
		},
		newUnstructured("group/version", "TheKind", "ns-foo", "name-foo"),
		newUnstructured("group2/version", "TheKind", "ns-foo", "name2-foo"),
		newUnstructured("group/version", "TheKind", "ns-foo1", "name-bar"),
		newUnstructuredReady("group/version", "TheKind", "ns-foo", "name-baz"),
		newUnstructured("group2/version", "TheKind", "ns-foo", "name2-baz"),
		newUnstructured("group/version", "AnotherKind", "ns-foo", "name2-baz"),
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

func TestScraperCantPullResource(t *testing.T) {
	rcConfig := []Resource{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}

	scheme := runtime.NewScheme()

	client := fake.NewSimpleDynamicClient(scheme, newUnstructured("group/version", "TheKind", "ns-foo", "name-foo"))

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

func newUnstructured(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"status": map[string]interface{}{
				"state": "ok",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "ConditionType1",
						"status":             "True",
						"reason":             "Reason1",
						"message":            "some message",
						"lastTransitionTime": "2023-09-01T15:46:59Z",
						"observedGeneration": "2",
					},
				},
			},
		},
	}
}

func newUnstructuredReady(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"status": map[string]interface{}{
				"state": "Ready",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "ConditionType1",
						"status":             "True",
						"reason":             "Reason1",
						"message":            "some message",
						"lastTransitionTime": "2023-09-01T15:46:59Z",
						"observedGeneration": "2",
					},
				},
			},
		},
	}
}
