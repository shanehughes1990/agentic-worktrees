package lifecycle

import (
	"errors"
	"fmt"
	"strings"

	"agentic-orchestrator/internal/domain/failures"
)

type ListenerType string

const (
	ListenerTypeGraphQL  ListenerType = "graphql"
	ListenerTypeInternal ListenerType = "internal"
	ListenerTypeWebhook  ListenerType = "webhook"
	ListenerTypeSlack    ListenerType = "slack"
	ListenerTypeBus      ListenerType = "bus"
)

type ListenerTarget struct {
	ListenerID   string
	ListenerType ListenerType
}

func (target ListenerTarget) Validate() error {
	if strings.TrimSpace(target.ListenerID) == "" {
		return failures.WrapTerminal(errors.New("listener_id is required"))
	}
	switch target.ListenerType {
	case ListenerTypeGraphQL, ListenerTypeInternal, ListenerTypeWebhook, ListenerTypeSlack, ListenerTypeBus:
		return nil
	default:
		return failures.WrapTerminal(fmt.Errorf("unsupported listener_type %q", target.ListenerType))
	}
}
