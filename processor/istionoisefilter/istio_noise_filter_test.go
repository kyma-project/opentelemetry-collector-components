package istionoisefilter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/metadata"
)

func TestIstioNoiseFilter(t *testing.T) {
	testCases := []struct {
		name              string
		spanAttrs         []map[string]any
		resourceAttrs     map[string]any
		expectedSpanCount int
	}{
		{
			name: "keeps span if not istio proxy",
			spanAttrs: []map[string]any{
				{"component": "not-proxy"},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 1,
		},
		{
			name: "drops telemetry module component span",
			spanAttrs: []map[string]any{
				{"component": "proxy", "istio.canonical_service": "telemetry-fluent-bit"},
			},
			resourceAttrs:     map[string]any{"k8s.namespace.name": "kyma-system"},
			expectedSpanCount: 0,
		},
		{
			name: "drops availability service probe span",
			spanAttrs: []map[string]any{
				{
					"component":               "proxy",
					"istio.canonical_service": "istio-ingressgateway",
					"http.method":             "GET",
					"http.url":                "https://healthz.foo/healthz/ready",
					"upstream_cluster.name":   "outbound|12345|svc|foo",
				},
			},
			resourceAttrs:     map[string]any{"k8s.namespace.name": "istio-system"},
			expectedSpanCount: 0,
		},
		{
			name: "drops log gateway span",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "POST",
					"http.url":              "https://telemetry-otlp-logs.kyma-system.svc:4317/v1/logs",
					"upstream_cluster.name": "outbound|4317|svc|telemetry-otlp-logs.kyma-system.svc.cluster.local",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 0,
		},
		{
			name: "drops metric gateway span",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "POST",
					"http.url":              "https://telemetry-otlp-metrics.kyma-system.svc:4317/v1/logs",
					"upstream_cluster.name": "outbound|4317|svc|telemetry-otlp-metrics.kyma-system.svc.cluster.local",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 0,
		},
		{
			name: "drops trace gateway span",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "POST",
					"http.url":              "https://telemetry-otlp-traces.kyma-system.svc:4317/v1/logs",
					"upstream_cluster.name": "outbound|4317|svc|telemetry-otlp-traces.kyma-system.svc.cluster.local",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 0,
		},
		{
			name: "drops VictoriaMetrics scrape span",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "GET",
					"user_agent":            "vm_promscrape/1.0",
					"upstream_cluster.name": "inbound|12345|svc|foo",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 0,
		},
		{
			name: "drops metric agent scrape span",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "GET",
					"user_agent":            "kyma-otelcol/1.0",
					"upstream_cluster.name": "inbound|12345|svc|foo",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 0,
		},
		{
			name: "keeps span if not matching any filter",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "GET",
					"user_agent":            "curl/7.68.0",
					"upstream_cluster.name": "inbound|12345|svc|foo",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 1,
		},
		{
			name: "mixed: one filtered, one kept",
			spanAttrs: []map[string]any{
				{
					"component":             "proxy",
					"http.method":           "POST",
					"http.url":              "https://telemetry-otlp-logs.kyma-system.svc:4317/v1/logs",
					"upstream_cluster.name": "outbound|4317|svc|telemetry-otlp-logs.kyma-system.svc.cluster.local",
				},
				{
					"component":             "proxy",
					"http.method":           "GET",
					"user_agent":            "curl/7.68.0",
					"upstream_cluster.name": "inbound|12345|svc|foo",
				},
			},
			resourceAttrs:     map[string]any{},
			expectedSpanCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			tp, err := factory.CreateTraces(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
			require.NoError(t, err)
			require.NotNil(t, tp)

			td := generateTraces(tc.resourceAttrs, tc.spanAttrs)
			err = tp.ConsumeTraces(t.Context(), td)
			require.NoError(t, err)
			require.Equal(t, tc.expectedSpanCount, td.SpanCount())
		})
	}
}

func generateTraces(resourceAttrs map[string]any, spanAttrs []map[string]any) ptrace.Traces {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	resource := rs.Resource()
	resource.Attributes().FromRaw(resourceAttrs)

	scopeSpans := rs.ScopeSpans().AppendEmpty()

	for _, attrs := range spanAttrs {
		span := scopeSpans.Spans().AppendEmpty()
		span.SetName("test-span")
		span.SetKind(ptrace.SpanKindServer)
		span.Attributes().FromRaw(attrs)
	}

	return traces
}
