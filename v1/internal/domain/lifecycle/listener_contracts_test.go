package lifecycle

import "testing"

func TestListenerTargetValidate(t *testing.T) {
	valid := ListenerTarget{ListenerID: "graphql_default", ListenerType: ListenerTypeGraphQL}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid listener target, got %v", err)
	}
	phaseTwo := []ListenerTarget{
		{ListenerID: "webhook_default", ListenerType: ListenerTypeWebhook},
		{ListenerID: "slack_default", ListenerType: ListenerTypeSlack},
		{ListenerID: "bus_default", ListenerType: ListenerTypeBus},
	}
	for _, listener := range phaseTwo {
		if err := listener.Validate(); err != nil {
			t.Fatalf("expected phase-2 listener target to validate (%s), got %v", listener.ListenerType, err)
		}
	}

	invalid := ListenerTarget{ListenerID: "", ListenerType: ListenerTypeInternal}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected validation error for empty listener id")
	}
}
