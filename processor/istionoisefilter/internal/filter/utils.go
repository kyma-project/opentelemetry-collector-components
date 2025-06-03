package filter

import "go.opentelemetry.io/collector/pdata/pcommon"

func getStringAttrOrEmpty(attrs pcommon.Map, key string) string {
	attr, ok := attrs.Get(key)
	if !ok {
		return ""
	}

	return attr.Str()
}
