package istionoisefilter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/istionoisefilter/internal/metadata"
)

func TestIstioNoiseFilter_Spans(t *testing.T) {
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
			name: "drops fluent-bit span",
			spanAttrs: []map[string]any{
				{"component": "proxy", "istio.canonical_service": "telemetry-fluent-bit"},
			},
			resourceAttrs:     map[string]any{"k8s.namespace.name": "kyma-system"},
			expectedSpanCount: 0,
		},
		{
			name: "drops metric gateway span",
			spanAttrs: []map[string]any{
				{"component": "proxy", "istio.canonical_service": "telemetry-metric-gateway"},
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
			name: "drops if user agent is vm_promscrape",
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
			name: "drops if user agent is kyma-otelcol",
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

func TestIstioNoiseFilter_Logs(t *testing.T) {
	testCases := []struct {
		name             string
		logAttrs         []map[string]any
		resourceAttrs    map[string]any
		expectedLogCount int
	}{
		{
			name: "keeps log if not istio module",
			logAttrs: []map[string]any{
				{
					"kyma.module": "other",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 1,
		},
		{
			name: "drops log if server address is log gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module":    "istio",
					"server.address": "telemetry-otlp-logs.kyma-system.svc:4317",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "drops log if server address is metric gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module":    "istio",
					"server.address": "telemetry-otlp-metrics.kyma-system.svc:4318",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "drops log if server address is trace gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module":    "istio",
					"server.address": "telemetry-otlp-traces.kyma-system.svc:4318",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by metric gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name":  "kyma-system",
				"k8s.deployment.name": "telemetry-metric-gateway",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by log gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name":  "kyma-system",
				"k8s.deployment.name": "telemetry-log-gateway",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by trace gateway",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name":  "kyma-system",
				"k8s.deployment.name": "telemetry-metric-gateway",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by log agent",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name": "kyma-system",
				"k8s.daemonset.name": "telemetry-log-agent",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by metric agent",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name": "kyma-system",
				"k8s.daemonset.name": "telemetry-log-agent",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if emitted by fluent bit",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs: map[string]any{
				"k8s.namespace.name": "kyma-system",
				"k8s.daemonset.name": "telemetry-fluent-bit",
			},
			expectedLogCount: 0,
		},
		{
			name: "drops log if vm_promscrape user agent",
			logAttrs: []map[string]any{
				{
					"kyma.module":         "istio",
					"http.request.method": "GET",
					"http.direction":      "inbound",
					"user_agent.original": "vm_promscrape/1.0",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "drops log if kyma-otelcol user agent",
			logAttrs: []map[string]any{
				{
					"kyma.module":         "istio",
					"http.request.method": "GET",
					"http.direction":      "inbound",
					"user_agent.original": "kyma-otelcol/1.2.3",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "drops if healthz domain and /healthz/ready path",
			logAttrs: []map[string]any{
				{
					"kyma.module":         "istio",
					"http.request.method": "GET",
					"http.direction":      "outbound",
					"server.address":      "healthz.foo.bar",
					"url.path":            "/healthz/ready",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 0,
		},
		{
			name: "keeps log healthz domain but wrong path",
			logAttrs: []map[string]any{
				{
					"kyma.module":         "istio",
					"http.request.method": "GET",
					"http.direction":      "outbound",
					"server.address":      "healthz.foo.bar",
					"url.path":            "/not/ready",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 1,
		},
		{
			name: "keeps log if kyma.module is istio but no other rule matches",
			logAttrs: []map[string]any{
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 1,
		},
		{
			name: "mixed: one dropped, one kept",
			logAttrs: []map[string]any{
				{
					"kyma.module":    "istio",
					"server.address": "telemetry-otlp-logs.kyma-system.svc:4317",
				},
				{
					"kyma.module": "istio",
				},
			},
			resourceAttrs:    map[string]any{},
			expectedLogCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			lp, err := factory.CreateLogs(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
			require.NoError(t, err)
			require.NotNil(t, lp)

			ld := generateLogs(tc.resourceAttrs, tc.logAttrs)
			err = lp.ConsumeLogs(t.Context(), ld)
			require.NoError(t, err)
			require.Equal(t, tc.expectedLogCount, ld.LogRecordCount())
		})
	}
}

func TestIstioNoiseFilter_Metrics(t *testing.T) {
	testCases := []struct {
		name         string
		metricName   string
		attrs        map[string]any
		metricType   pmetric.MetricType
		expectedDrop bool
	}{
		{
			name:         "keeps non-istio metric (Gauge)",
			metricName:   "custom.metric",
			attrs:        map[string]any{"source_workload": "telemetry-metric-agent"},
			metricType:   pmetric.MetricTypeGauge,
			expectedDrop: false,
		},
		{
			name:         "drops istio metric with source_workload telemetry-metric-agent (Gauge)",
			metricName:   "istio_requests_total",
			attrs:        map[string]any{"source_workload": "telemetry-metric-agent"},
			metricType:   pmetric.MetricTypeGauge,
			expectedDrop: true,
		},
		{
			name:         "drops istio metric with source_workload telemetry-metric-agent (Sum)",
			metricName:   "istio_bytes_sent",
			attrs:        map[string]any{"source_workload": "telemetry-metric-agent"},
			metricType:   pmetric.MetricTypeSum,
			expectedDrop: true,
		},
		{
			name:         "drops istio metric with destination_workload telemetry-log-gateway (Histogram)",
			metricName:   "istio_latency",
			attrs:        map[string]any{"destination_workload": "telemetry-log-gateway"},
			metricType:   pmetric.MetricTypeHistogram,
			expectedDrop: true,
		},
		{
			name:         "drops istio metric with destination_workload telemetry-metric-gateway (ExponentialHistogram)",
			metricName:   "istio_latency_exp",
			attrs:        map[string]any{"destination_workload": "telemetry-metric-gateway"},
			metricType:   pmetric.MetricTypeExponentialHistogram,
			expectedDrop: true,
		},
		{
			name:         "drops istio metric with destination_workload telemetry-trace-gateway (Summary)",
			metricName:   "istio_summary",
			attrs:        map[string]any{"destination_workload": "telemetry-trace-gateway"},
			metricType:   pmetric.MetricTypeSummary,
			expectedDrop: true,
		},
		{
			name:         "keeps istio metric with other destination_workload (Gauge)",
			metricName:   "istio.requests.total",
			attrs:        map[string]any{"destination_workload": "user-app"},
			metricType:   pmetric.MetricTypeGauge,
			expectedDrop: false,
		},
		{
			name:         "does not drop istio metric with no relevant attributes (Sum)",
			metricName:   "istio.requests.total",
			attrs:        map[string]any{"foo": "bar"},
			metricType:   pmetric.MetricTypeSum,
			expectedDrop: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := generateMetrics(tc.metricName, tc.attrs, tc.metricType)
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			mp, err := factory.CreateMetrics(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
			require.NoError(t, err)
			require.NotNil(t, mp)

			err = mp.ConsumeMetrics(t.Context(), metrics)
			require.NoError(t, err)

			got := metrics.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
			if tc.expectedDrop {
				require.Equal(t, 0, got.Len())
			} else {
				require.Equal(t, 1, got.Len())
			}
		})
	}
}

func generateTraces(resourceAttrs map[string]any, spanAttrs []map[string]any) ptrace.Traces {
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	resource := rs.Resource()
	resource.Attributes().FromRaw(resourceAttrs)

	scopeSpans := rs.ScopeSpans().AppendEmpty()

	for i, attrs := range spanAttrs {
		span := scopeSpans.Spans().AppendEmpty()
		span.SetName(fmt.Sprintf("test-span-%d", i))
		span.SetKind(ptrace.SpanKindServer)
		span.Attributes().FromRaw(attrs)
	}

	return traces
}

func generateLogs(resourceAttrs map[string]any, logAttrs []map[string]any) plog.Logs {
	logs := plog.NewLogs()
	resLogs := logs.ResourceLogs().AppendEmpty()
	resource := resLogs.Resource()
	resource.Attributes().FromRaw(resourceAttrs)

	scopeLogs := resLogs.ScopeLogs().AppendEmpty()

	for i, attrs := range logAttrs {
		logRecord := scopeLogs.LogRecords().AppendEmpty()
		logRecord.Body().SetStr(fmt.Sprintf("test-log-%d", i))
		logRecord.Attributes().FromRaw(attrs)
	}

	return logs
}

func generateMetrics(metricName string, dataPointAttrs map[string]any, metricType pmetric.MetricType) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(metricName)
	switch metricType {
	case pmetric.MetricTypeGauge:
		metric.SetEmptyGauge()
		dp := metric.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().FromRaw(dataPointAttrs)
	case pmetric.MetricTypeSum:
		metric.SetEmptySum()
		dp := metric.Sum().DataPoints().AppendEmpty()
		dp.Attributes().FromRaw(dataPointAttrs)
	case pmetric.MetricTypeHistogram:
		metric.SetEmptyHistogram()
		dp := metric.Histogram().DataPoints().AppendEmpty()
		dp.Attributes().FromRaw(dataPointAttrs)
	case pmetric.MetricTypeExponentialHistogram:
		metric.SetEmptyExponentialHistogram()
		dp := metric.ExponentialHistogram().DataPoints().AppendEmpty()
		dp.Attributes().FromRaw(dataPointAttrs)
	case pmetric.MetricTypeSummary:
		metric.SetEmptySummary()
		dp := metric.Summary().DataPoints().AppendEmpty()
		dp.Attributes().FromRaw(dataPointAttrs)
	}
	return metrics
}
