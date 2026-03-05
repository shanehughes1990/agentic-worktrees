package resolvers

import (
	"agentic-orchestrator/internal/interface/graphql/models"
	"strings"
)

func graphErrorFromError(err error) models.GraphError {
	if err == nil {
		return models.GraphError{Code: models.GraphErrorCodeInternal, Message: "unknown error"}
	}
	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "required"), strings.Contains(lower, "invalid"):
		return models.GraphError{Code: models.GraphErrorCodeValidation, Message: message}
	case strings.Contains(lower, "not found"):
		return models.GraphError{Code: models.GraphErrorCodeNotFound, Message: message}
	case strings.Contains(lower, "forbidden"):
		return models.GraphError{Code: models.GraphErrorCodeForbidden, Message: message}
	case strings.Contains(lower, "unauthorized"):
		return models.GraphError{Code: models.GraphErrorCodeUnauthorized, Message: message}
	case strings.Contains(lower, "duplicate"), strings.Contains(lower, "already"):
		return models.GraphError{Code: models.GraphErrorCodeConflict, Message: message}
	default:
		return models.GraphError{Code: models.GraphErrorCodeInternal, Message: message}
	}
}
