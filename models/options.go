package models

import (
	"github.com/joaosoft/logger"
	"github.com/joaosoft/manager"
)

// EventSourcingOption ...
type EventSourcingOption func(es *EventSourcing)

// Reconfigure ...
func (es *EventSourcing) Reconfigure(options ...EventSourcingOption) {
	for _, option := range options {
		option(es)
	}
}

// WithConfiguration ...
func WithConfiguration(config *EventSourcingConfig) EventSourcingOption {
	return func(es *EventSourcing) {
		es.config = config
	}
}

// WithLogger ...
func WithLogger(l logger.ILogger) EventSourcingOption {
	return func(es *EventSourcing) {
		es.logger = l
		es.isLogExternal = true
	}
}

// WithLogLevel ...
func WithLogLevel(level logger.Level) EventSourcingOption {
	return func(es *EventSourcing) {
		es.logger.SetLevel(level)
	}
}

// WithManager ...
func WithManager(mgr *manager.Manager) EventSourcingOption {
	return func(es *EventSourcing) {
		es.pm = mgr
	}
}
