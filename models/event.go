package models

import (
	"time"

	"github.com/oklog/ulid"
)

type IEvent interface {
	GetId() ulid.ULID
	GetName() string
	GetData() interface{}
	GetCreatedAt() time.Time
}

type Event struct {
	Id        ulid.ULID
	Name      string
	Data      interface{}
	CreatedAt time.Time
}

func (event *Event) GetId() ulid.ULID {
	return event.Id
}

func (event *Event) GetName() string {
	return event.Name
}

func (event *Event) GetData() interface{} {
	return event.Data
}

func (event *Event) GetCreatedAt() time.Time {
	return event.CreatedAt
}
