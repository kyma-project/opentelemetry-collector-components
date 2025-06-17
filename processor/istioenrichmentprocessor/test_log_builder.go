package istioenrichmentprocessor

import "go.opentelemetry.io/collector/pdata/plog"

type PLogBuilder struct {
	resourceAttributes map[string]string
	scopeAttributes    map[string]string
	logAttributes      map[string]string
	scopeName          string
	scopeVersion       string
}

func NewPLogBuilder() *PLogBuilder {
	return &PLogBuilder{
		resourceAttributes: make(map[string]string),
		scopeAttributes:    make(map[string]string),
		logAttributes:      make(map[string]string),
	}
}

func (b *PLogBuilder) WithResourceAttributes(attrs map[string]string) *PLogBuilder {
	b.resourceAttributes = attrs
	return b
}

func (b *PLogBuilder) WithScopeAttributes(attrs map[string]string) *PLogBuilder {
	b.scopeAttributes = attrs
	return b
}

func (b *PLogBuilder) WithLogAttributes(attrs map[string]string) *PLogBuilder {
	b.logAttributes = attrs
	return b
}

func (b *PLogBuilder) WithScopeName(name string) *PLogBuilder {
	b.scopeName = name
	return b
}

func (b *PLogBuilder) WithScopeVersion(version string) *PLogBuilder {
	b.scopeVersion = version
	return b
}

func (b *PLogBuilder) Build() plog.Logs {
	logs := plog.NewLogs()

	resLogs := logs.ResourceLogs().AppendEmpty()
	for k, v := range b.resourceAttributes {
		resLogs.Resource().Attributes().PutStr(k, v)
	}

	scopeLogs := resLogs.ScopeLogs().AppendEmpty()
	for k, v := range b.scopeAttributes {
		scopeLogs.Scope().Attributes().PutStr(k, v)
	}
	scopeLogs.Scope().SetName(b.scopeName)
	scopeLogs.Scope().SetVersion(b.scopeVersion)

	logRecord := scopeLogs.LogRecords().AppendEmpty()
	for k, v := range b.logAttributes {
		logRecord.Attributes().PutStr(k, v)
	}
	return logs
}
