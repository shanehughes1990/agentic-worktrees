package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

const RedactionPolicyVersion = "v1"

var sensitiveKeyPattern = regexp.MustCompile(`(?i)(token|password|secret|api[_-]?key|authorization|cookie)`)

func RedactPayload(payload map[string]any) (map[string]any, bool) {
	if payload == nil {
		return map[string]any{"_redaction_policy_version": RedactionPolicyVersion}, false
	}
	redacted := false
	cloned := redactMap(payload, &redacted)
	cloned["_redaction_policy_version"] = RedactionPolicyVersion
	if redacted {
		cloned["_redacted"] = true
	}
	return cloned, redacted
}

func redactMap(source map[string]any, redacted *bool) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		if sensitiveKeyPattern.MatchString(strings.TrimSpace(key)) {
			*redacted = true
			result[key] = redactedDigest(value)
			continue
		}
		result[key] = redactValue(value, redacted)
	}
	return result
}

func redactValue(value any, redacted *bool) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMap(typed, redacted)
	case []any:
		items := make([]any, 0, len(typed))
		for _, entry := range typed {
			items = append(items, redactValue(entry, redacted))
		}
		return items
	default:
		return value
	}
}

func redactedDigest(value any) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", value)))
	return "[REDACTED_SHA256:" + hex.EncodeToString(hash[:8]) + "]"
}
