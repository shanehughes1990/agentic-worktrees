package security

import "testing"

func TestRedactPayloadRedactsSensitiveKeysRecursively(t *testing.T) {
	payload, redacted := RedactPayload(map[string]any{
		"token": "super-secret-token",
		"nested": map[string]any{
			"password": "p@ssw0rd",
			"safe":     "value",
		},
	})
	if !redacted {
		t.Fatalf("expected redacted=true")
	}
	if payload["_redaction_policy_version"] != RedactionPolicyVersion {
		t.Fatalf("expected redaction policy version stamp")
	}
	if payload["token"] == "super-secret-token" {
		t.Fatalf("expected token to be redacted")
	}
	nested, ok := payload["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map")
	}
	if nested["password"] == "p@ssw0rd" {
		t.Fatalf("expected nested password to be redacted")
	}
	if nested["safe"] != "value" {
		t.Fatalf("expected safe field to remain unchanged")
	}
}
