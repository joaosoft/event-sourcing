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

func (e *Event) GetId() ulid.ULID {
	return e.Id
}

func (e *Event) GetName() string {
	return e.Name
}

func (e *Event) GetData() interface{} {
	return e.Data
}

func (e *Event) GetCreatedAt() time.Time {
	return e.CreatedAt
}
