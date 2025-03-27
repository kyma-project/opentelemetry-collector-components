package serviceenrichmentprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/kyma-project/opentelemetry-collector-components/processor/serviceenrichmentprocessor/internal/metadata"
)

func TestProcessLogs(t *testing.T) {
	tt := []struct {
		name                string
		logs                plog.Logs
		expectedServiceName string
	}{
		{
			name: "logs with service name not set and k8s-io-app-name-set",
			logs: setLogs(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name not set and app-name-set",
			logs: setLogs(map[string]string{
				"kyma.app_name": "foo-app-name",
			}),
			expectedServiceName: "foo-app-name",
		},
		{
			name: "logs with service name not set and deployment name set",
			logs: setLogs(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expectedServiceName: "foo-deployment-name",
		},
		{
			name: "logs with service name not set and daemonset name set",
			logs: setLogs(map[string]string{
				"k8s.daemonset.name": "foo-daemonset-name",
			}),
			expectedServiceName: "foo-daemonset-name",
		},
		{
			name: "logs with service name not set and job name is set",
			logs: setLogs(map[string]string{
				"k8s.job.name": "foo-job-name",
			}),
			expectedServiceName: "foo-job-name",
		},
		{
			name: "logs with service name set to unknown_service",
			logs: setLogs(map[string]string{
				"service.name":                "unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "foo-k8s-io-app-name",
		},
		{
			name: "logs with service name set to test_unknown_service",
			logs: setLogs(map[string]string{
				"service.name":                "test_unknown_service",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "test_unknown_service",
		},
		{
			name: "logs with service name set to unknown_service_test",
			logs: setLogs(map[string]string{
				"service.name":                "unknown_service_test",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service_test",
		},
		{
			name: "logs with service name set to unknown_service:",
			logs: setLogs(map[string]string{
				"service.name":                "unknown_service:",
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expectedServiceName: "unknown_service:",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sink := new(consumertest.LogsSink)
			config := Config{
				CustomLabels: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}

			factory := NewFactory()
			cm, err := factory.CreateLogs(
				t.Context(),
				processortest.NewNopSettings(metadata.Type),
				config,
				sink,
			)
			require.NoError(t, err)
			require.NotNil(t, cm)

			err = cm.Start(t.Context(), componenttest.NewNopHost())
			require.NoError(t, err)

			cErr := cm.ConsumeLogs(t.Context(), tc.logs)
			require.NoError(t, cErr)

			got := sink.AllLogs()
			require.Len(t, got, 1)
			for _, l := range got {
				for i := 0; i < l.ResourceLogs().Len(); i++ {
					attr := l.ResourceLogs().At(i).Resource().Attributes()
					svcName, ok := attr.Get("service.name")
					require.True(t, ok)
					require.Equal(t, tc.expectedServiceName, svcName.AsString())
				}
			}
		})
	}
}

func setLogs(attrs ...map[string]string) plog.Logs {
	logs := plog.NewLogs()
	for _, attr := range attrs {
		resLogs := logs.ResourceLogs().AppendEmpty()
		for k, v := range attr {
			resLogs.Resource().Attributes().PutStr(k, v)
		}
	}
	return logs
}
