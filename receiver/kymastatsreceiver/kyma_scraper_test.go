package kymastatsreceiver

import (
	"context"
	"errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

	statusNoCondition = map[string]interface{}{
		"state": "Ready",
	}

	statusNoConditionType = map[string]interface{}{
		"state": "Ready",
		"conditions": []interface{}{
			map[string]interface{}{
				"status":  "False",
				"reason":  "Reason1",
				"message": "some message",
			},
		},
	}

	statusNoConditionStatus = map[string]interface{}{
		"state": "Ready",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "ConditionType1",
				"reason":  "Reason1",
				"message": "some message",
			},
		},
	}

	statusNoConditionReason = map[string]interface{}{
		"state": "Ready",
		"conditions": []interface{}{
			map[string]interface{}{
				"type":    "ConditionType1",
				"status":  "True",
				"message": "some message",
			},
		},
	}
)

func TestScraper(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
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

func TestScraperCantPullResource(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
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

func TestScraperResourceWithoutStatus(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}
	scheme := runtime.NewScheme()
	obj := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")

	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.NoError(t, err)
}

func TestScraperResourceWithNoCondition(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}
	scheme := runtime.NewScheme()
	obj := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")
	unstructured.SetNestedMap(obj, statusNoCondition, "status")
	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.NoError(t, err)
}

func TestScraperResourceNoConditionType(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}
	scheme := runtime.NewScheme()
	obj := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")
	unstructured.SetNestedMap(obj, statusNoConditionType, "status")
	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.NoError(t, err)
}

func TestScraperResourceNoConditionStatus(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}
	scheme := runtime.NewScheme()
	obj := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")
	unstructured.SetNestedMap(obj, statusNoConditionStatus, "status")
	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.NoError(t, err)
}

func TestScraperResourceNoConditionReason(t *testing.T) {
	rcConfig := []ModuleResourceConfig{
		{
			ResourceGroup:   "group",
			ResourceName:    "thekinds",
			ResourceVersion: "version",
		},
	}
	scheme := runtime.NewScheme()
	obj := newUnstructuredObject("group/version", "TheKind", "ns-foo", "name-foo")
	unstructured.SetNestedMap(obj, statusNoConditionReason, "status")
	client := fake.NewSimpleDynamicClient(scheme, &unstructured.Unstructured{Object: obj})

	r, err := newKymaScraper(
		client,
		receivertest.NewNopSettings(),
		rcConfig,
		metadata.DefaultMetricsBuilderConfig(),
	)

	require.NoError(t, err)

	_, err = r.Scrape(context.Background())
	require.NoError(t, err)
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
