package models

import (
	"time"
)

type EventsType map[string]func(*Aggregate, IEvent) error
type Aggregate struct {
	Id        string
	Type      string
	Version   int64
	Data      interface{}
	Events    []IEvent
	CreatedAt time.Time
	UpdatedAt time.Time
	events    EventsType
}

func NewAggregate(id string, typ string, data interface{}) *Aggregate {
	return &Aggregate{
		Id:     id,
		Type:   typ,
		Data:   data,
		events: make(EventsType),
	}
}

func (a *Aggregate) AddEventHandler(name string, function func(*Aggregate, IEvent) error) {
	a.events[name] = function
}

func (a *Aggregate) Causes(event IEvent) error {
	a.addEvent(event)
	return a.apply(event)
}

func (a *Aggregate) addEvent(event IEvent) {
	a.Events = append(a.Events, event)
}

func (a *Aggregate) apply(event IEvent) error {

	for name, handler := range a.events {
		if event.GetName() == name {
			return handler(a, event)
		}
	}

	return nil
}
