package istioenrichmentprocessor

import (
	"context"
	"net"
	"strings"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

const (
	kymaModuleAttributeName             = "kyma.module"
	kymaModuleAttributeValue            = "istio"
	istioScopeName                      = "io.kyma-project.telemetry/istio"
	clientAddressAttributeName          = "client.address"
	clientPortAttributeName             = "client.port"
	networkProtocolNameAttributeName    = "network.protocol.name"
	networkProtocolVersionAttributeName = "network.protocol.version"
	defaultSeverityText                 = "INFO"
	defaultSeverityNumber               = plog.SeverityNumberInfo
)

type istioEnrichmentProcessor struct {
	logger *zap.Logger
	config Config
}

func newIstioEnrichmentProcessor(logger *zap.Logger, cfg Config) *istioEnrichmentProcessor {
	return &istioEnrichmentProcessor{
		logger: logger,
		config: cfg,
	}
}

func (iep *istioEnrichmentProcessor) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	resourceLogs := logs.ResourceLogs()
	for _, r := range resourceLogs.All() {
		for _, s := range r.ScopeLogs().All() {
			updateScopeAttributes := true

			for _, l := range s.LogRecords().All() {
				moduleName, exist := l.Attributes().Get(kymaModuleAttributeName)
				if !exist || moduleName.Str() != kymaModuleAttributeValue {
					// If the log record does not have the kyma.module attribute set to "istio",
					// we skip the enrichment for this log record.
					continue
				}
				enrichSeverityAttributes(l)
				setNetworkProtocolAttributes(l)
				setClientAddressAttributes(l)
				l.Attributes().Remove(kymaModuleAttributeName)

				if updateScopeAttributes {
					iep.setScopeAttributes(s)
					updateScopeAttributes = false
				}
			}
		}
	}
	return logs, nil
}

func enrichSeverityAttributes(logR plog.LogRecord) {
	logR.SetSeverityText(defaultSeverityText)
	logR.SetSeverityNumber(defaultSeverityNumber)
}

func (iep *istioEnrichmentProcessor) setScopeAttributes(scopeLog plog.ScopeLogs) {
	scopeLog.Scope().SetName(istioScopeName)
	scopeLog.Scope().SetVersion(iep.config.ScopeVersion)
}

func setNetworkProtocolAttributes(logR plog.LogRecord) {
	networkProtocol, exist := logR.Attributes().Get(networkProtocolNameAttributeName)
	if exist && networkProtocol.Str() != "" {
		parts := strings.Split(networkProtocol.Str(), "/")
		if len(parts) == 2 {
			logR.Attributes().PutStr(networkProtocolNameAttributeName, parts[0])
			logR.Attributes().PutStr(networkProtocolVersionAttributeName, parts[1])
		}
	}
}

func setClientAddressAttributes(logR plog.LogRecord) {
	clientAddress, exist := logR.Attributes().Get(clientAddressAttributeName)
	if exist && clientAddress.Str() != "" {
		host, port, err := net.SplitHostPort(clientAddress.Str())
		if err == nil {
			logR.Attributes().PutStr(clientAddressAttributeName, host)
			logR.Attributes().PutStr(clientPortAttributeName, port)
		}
	}
}
