package models

import (
	"time"
)

type Aggregate struct {
	Id        string
	Type      string
	Version   int64
	Data      interface{}
	Events    []IEvent
	CreatedAt time.Time
	UpdatedAt time.Time
	events    map[string]func(*Aggregate, IEvent) error
}

func NewAggregate(id string, typ string, data interface{}) *Aggregate {
	return &Aggregate{
		Id:     id,
		Type:   typ,
		Data:   data,
		events: make(map[string]func(*Aggregate, IEvent) error),
	}
}

func (aggregate *Aggregate) AddEventHandler(name string, function func(*Aggregate, IEvent) error) {
	aggregate.events[name] = function
}

func (aggregate *Aggregate) Causes(event IEvent) {
	aggregate.addEvent(event)
	aggregate.apply(event)
}

func (aggregate *Aggregate) addEvent(event IEvent) {
	aggregate.Events = append(aggregate.Events, event)
}

func (aggregate *Aggregate) apply(event IEvent) error {

	for name, handler := range aggregate.events {
		if event.GetName() == name {
			return handler(aggregate, event)
		}
	}

	return nil
}
