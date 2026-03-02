package observability

import (
	domainobservability "agentic-orchestrator/internal/domain/shared/observability"
	"github.com/sirupsen/logrus"
)

type Entry struct {
	entry *logrus.Entry
}

func newEntry(entry *logrus.Entry) *Entry {
	if entry == nil {
		base := logrus.New()
		entry = logrus.NewEntry(base)
	}
	return &Entry{entry: entry}
}

func (entry *Entry) WithField(key string, value any) domainobservability.Entry {
	if entry == nil || entry.entry == nil {
		return entry
	}
	return newEntry(entry.entry.WithField(key, value))
}

func (entry *Entry) WithFields(fields map[string]any) domainobservability.Entry {
	if entry == nil || entry.entry == nil {
		return entry
	}
	return newEntry(entry.entry.WithFields(logrus.Fields(fields)))
}

func (entry *Entry) WithError(err error) domainobservability.Entry {
	if entry == nil || entry.entry == nil {
		return entry
	}
	return newEntry(entry.entry.WithError(err))
}

func (entry *Entry) Debug(message string) {
	if entry == nil || entry.entry == nil {
		return
	}
	entry.entry.Debug(message)
}

func (entry *Entry) Info(message string) {
	if entry == nil || entry.entry == nil {
		return
	}
	entry.entry.Info(message)
}

func (entry *Entry) Warn(message string) {
	if entry == nil || entry.entry == nil {
		return
	}
	entry.entry.Warn(message)
}

func (entry *Entry) Error(message string) {
	if entry == nil || entry.entry == nil {
		return
	}
	entry.entry.Error(message)
}

var _ domainobservability.Entry = (*Entry)(nil)
