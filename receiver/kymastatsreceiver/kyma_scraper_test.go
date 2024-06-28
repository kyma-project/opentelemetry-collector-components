package kymastatsreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

const dataLen = 6

func TestScraper(t *testing.T) {
	rcConfig := []internal.Resource{
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
		newUnstructured("group/version", "TheKind", "ns-foo", "name-bar"),
		newUnstructured("group/version", "TheKind", "ns-foo", "name-baz"),
		newUnstructured("group2/version", "TheKind", "ns-foo", "name2-baz"),
	)

	r, err := newKymaScraper(
		client,
		receivertest.NewNopCreateSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	md, err := r.Scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, dataLen, md.DataPointCount())
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
