package models

import (
	"github.com/joaosoft/logger"
	"github.com/joaosoft/manager"
)

// EventSourcingOption ...
type EventSourcingOption func(client *EventSourcing)

// Reconfigure ...
func (eventSourving *EventSourcing) Reconfigure(options ...EventSourcingOption) {
	for _, option := range options {
		option(eventSourving)
	}
}

// WithConfiguration ...
func WithConfiguration(config *EventSourcingConfig) EventSourcingOption {
	return func(client *EventSourcing) {
		client.config = config
	}
}

// WithLogger ...
func WithLogger(l logger.ILogger) EventSourcingOption {
	return func(eventSourving *EventSourcing) {
		eventSourving.logger = l
		eventSourving.isLogExternal = true
	}
}

// WithLogLevel ...
func WithLogLevel(level logger.Level) EventSourcingOption {
	return func(eventSourving *EventSourcing) {
		eventSourving.logger.SetLevel(level)
	}
}

// WithManager ...
func WithManager(mgr *manager.Manager) EventSourcingOption {
	return func(eventSourving *EventSourcing) {
		eventSourving.pm = mgr
	}
}
