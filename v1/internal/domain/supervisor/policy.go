package supervisor

import (
	"cmp"
	"context"
	"fmt"
	"slices"
)

type Rule interface {
	Name() string
	Priority() int
	Evaluate(ctx context.Context, state State, signal Signal) (Decision, bool, error)
}

func Evaluate(ctx context.Context, state State, signal Signal, rules []Rule) (Decision, error) {
	if err := signal.Validate(); err != nil {
		return Decision{}, err
	}
	if len(rules) == 0 {
		return Decision{}, ErrNoMatchingRule
	}
	sortedRules := append([]Rule(nil), rules...)
	slices.SortStableFunc(sortedRules, func(left Rule, right Rule) int {
		if left.Priority() == right.Priority() {
			return cmp.Compare(left.Name(), right.Name())
		}
		return cmp.Compare(right.Priority(), left.Priority())
	})

	for _, rule := range sortedRules {
		decision, matched, err := rule.Evaluate(ctx, state, signal)
		if err != nil {
			return Decision{}, fmt.Errorf("evaluate rule %q: %w", rule.Name(), err)
		}
		if !matched {
			continue
		}
		decision.CorrelationIDs = signal.CorrelationIDs
		decision.SignalType = signal.Type
		decision.FromState = state
		decision.RuleName = rule.Name()
		decision.RulePriority = rule.Priority()
		decision.OccurredAt = signal.OccurredAt
		decision.Attempt = signal.Attempt
		decision.MaxRetry = signal.MaxRetry
		decision.FailureClass = signal.FailureClass
		decision.AttentionZone = signal.AttentionZone
		decision.Metadata = signal.Metadata
		if err := decision.Validate(); err != nil {
			return Decision{}, err
		}
		return decision, nil
	}
	return Decision{}, ErrNoMatchingRule
}
