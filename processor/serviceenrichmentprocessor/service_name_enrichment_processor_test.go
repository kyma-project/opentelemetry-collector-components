package serviceenrichmentprocessor

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

func TestFetchFirstAvailableServiceName(t *testing.T) {
	tt := []struct {
		name     string
		attr     pcommon.Map
		expected string
	}{
		{
			name:     "empty attributes",
			attr:     setAttributes(),
			expected: "unknown_service",
		},
		{
			name: "kyma.kubernetes_app_io_name is set",
			attr: setAttributes(map[string]string{
				"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
			}),
			expected: "foo-k8s-io-app-name",
		},
		{
			name: "when multiple attributes are set, the first one is returned",
			attr: setAttributes(
				map[string]string{
					"kyma.kubernetes_io_app_name": "foo-k8s-io-app-name",
					"kyma.app_name":               "foo-app-name",
				}),
			expected: "foo-k8s-io-app-name",
		},
		{
			name: "kyma.app_name is set",
			attr: setAttributes(map[string]string{
				"kyma.app_name": "foo-app-name",
			}),
			expected: "foo-app-name",
		},
		{
			name: "when multiple attributes are set, if app_name and deployment_name are set, app_name is returned",
			attr: setAttributes(
				map[string]string{
					"kyma.app_name":       "foo-app-name",
					"k8s.deployment.name": "foo-deployment-name",
				}),
			expected: "foo-app-name",
		},
		{
			name: "k8s.deployment_name is set",
			attr: setAttributes(map[string]string{
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expected: "foo-deployment-name",
		},
		{
			name: "k8s.daemonset_name is set",
			attr: setAttributes(map[string]string{
				"k8s.daemonset.name": "foo-daemonset-name",
			}),
			expected: "foo-daemonset-name",
		},
		{
			name: "k8s.statefulset_name is set",
			attr: setAttributes(map[string]string{
				"k8s.statefulset.name": "foo-statefulset-name",
			}),
			expected: "foo-statefulset-name",
		},
		{
			name: "k8s.job_name is set",
			attr: setAttributes(map[string]string{
				"k8s.job.name": "foo-job-name",
			}),
			expected: "foo-job-name",
		},
		{
			name: "k8s.pod_name is set",
			attr: setAttributes(map[string]string{
				"k8s.pod.name": "foo-pod-name",
			}),
			expected: "foo-pod-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop()
			config := Config{
				additionalResourceAttributes: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			sep := newServiceEnrichmentProcessor(logger, config)
			got := sep.fetchFirstAvailableServiceName(tc.attr)
			if got != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got)
			}
		})
	}
}

func TestSetServiceName(t *testing.T) {
	tt := []struct {
		name     string
		attr     pcommon.Map
		expected string
	}{
		{
			name: "service name is set",
			attr: setAttributes(map[string]string{
				"service.name": "foo-service-name",
			}),
			expected: "foo-service-name",
		},
		{
			name: "service name is set and deployment name is set",
			attr: setAttributes(map[string]string{
				"service.name":        "foo-service-name",
				"k8s.deployment.name": "foo-deployment-name",
			}),
			expected: "foo-service-name",
		},
		{
			name: "service name is unknown but pod name is set",
			attr: setAttributes(map[string]string{
				"k8s.pod.name": "foo-pod-name",
			}),
			expected: "foo-pod-name",
		},
		{
			name: "service name is unset and app name is set",
			attr: setAttributes(map[string]string{
				"kyma.app_name": "foo-app-name",
			}),
			expected: "foo-app-name",
		},
		{
			name: "service name is empty and job name is set",
			attr: setAttributes(map[string]string{
				"service.name": "",
				"k8s.job.name": "foo-job-name",
			}),
			expected: "foo-job-name",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop()
			config := Config{
				additionalResourceAttributes: []string{
					"kyma.kubernetes_io_app_name",
					"kyma.app_name",
				},
			}
			sep := newServiceEnrichmentProcessor(logger, config)
			sep.enrichServiceName(tc.attr)
			got, _ := tc.attr.Get("service.name")
			if got.AsString() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got.AsString())
			}
		})
	}

}

func setAttributes(attrs ...map[string]string) pcommon.Map {
	attrMap := pcommon.NewMap()
	for _, attr := range attrs {
		for k, v := range attr {
			attrMap.PutStr(k, v)
		}
	}
	return attrMap
}
