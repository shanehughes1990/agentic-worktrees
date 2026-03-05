//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestTaskboardSchemaIncludesIngestionAudits(t *testing.T) {
	endpoint := strings.TrimSpace(os.Getenv("E2E_GRAPHQL_HTTP_ENDPOINT"))
	if endpoint == "" {
		endpoint = "http://localhost:8080/query"
	}

	query := `query TaskboardSchemaIntrospection {
  taskboardType: __type(name: "Taskboard") {
    fields { name }
  }
	taskType: __type(name: "TaskboardTask") {
		fields { name }
	}
  auditType: __type(name: "TaskModelAudit") {
    fields { name }
  }
}`
	body, err := json.Marshal(map[string]any{"query": query})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post graphql: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if errs, ok := payload["errors"].([]any); ok && len(errs) > 0 {
		t.Fatalf("graphql errors: %+v", errs)
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data payload")
	}

	taskboardType, ok := data["taskboardType"].(map[string]any)
	if !ok {
		t.Fatalf("missing taskboard type")
	}
	fields, ok := taskboardType["fields"].([]any)
	if !ok {
		t.Fatalf("missing taskboard fields")
	}
	foundIngestionAudits := false
	for _, field := range fields {
		entry, ok := field.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(fmt.Sprintf("%v", entry["name"])) == "ingestionAudits" {
			foundIngestionAudits = true
			break
		}
	}
	if !foundIngestionAudits {
		t.Fatalf("taskboard.ingestionAudits field not found")
	}

	taskType, ok := data["taskType"].(map[string]any)
	if !ok {
		t.Fatalf("missing taskboard task type")
	}
	taskFields, ok := taskType["fields"].([]any)
	if !ok {
		t.Fatalf("missing taskboard task fields")
	}
	foundTaskAudits := false
	for _, field := range taskFields {
		entry, ok := field.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(fmt.Sprintf("%v", entry["name"])) == "audits" {
			foundTaskAudits = true
			break
		}
	}
	if !foundTaskAudits {
		t.Fatalf("taskboardTask.audits field not found")
	}

	auditType, ok := data["auditType"].(map[string]any)
	if !ok {
		t.Fatalf("missing TaskModelAudit type")
	}
	auditFields, ok := auditType["fields"].([]any)
	if !ok {
		t.Fatalf("missing TaskModelAudit fields")
	}
	required := map[string]bool{
		"modelProvider":  false,
		"modelName":      false,
		"modelRunID":     false,
		"agentSessionID": false,
		"agentStreamID":  false,
	}
	for _, field := range auditFields {
		entry, ok := field.(map[string]any)
		if !ok {
			continue
		}
		name := strings.TrimSpace(fmt.Sprintf("%v", entry["name"]))
		if _, exists := required[name]; exists {
			required[name] = true
		}
	}
	for field, exists := range required {
		if !exists {
			t.Fatalf("TaskModelAudit.%s field not found", field)
		}
	}
}
