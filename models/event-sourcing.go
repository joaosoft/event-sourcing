package models

import (
	"event-sourcing/common"
	logger "github.com/joaosoft/logger"
	mapper "github.com/joaosoft/mapper"
	"reflect"
	"strings"
)

type IStorage interface {
	GetAggregate(id, typ string, obj interface{}) (aggregate *Aggregate, err error)
	StoreAggregate(aggregate *Aggregate) (err error)
}

type EventSourcing struct {
	storage IStorage
}

func NewEventSourcing(storage IStorage) *EventSourcing {
	return &EventSourcing{
		storage: storage,
	}
}
func (eventsourcing *EventSourcing) Save(aggregate *Aggregate) (err error) {

	if len(aggregate.Events) == 0 {
		obj := reflect.New(reflect.TypeOf(aggregate.Data).Elem()).Interface()
		oldAggregate, err := eventsourcing.storage.GetAggregate(aggregate.Id, aggregate.Type, obj)
		if err != nil {
			logger.WithField("error", err.Error()).Error("error saving aggregate")
			return err
		}

		aggregate.Events, err = getAggregateEventsByMapping(oldAggregate, aggregate)
		if err != nil {
			logger.WithField("error", err.Error()).Error("error saving aggregate")
			return err
		}
	}

	return eventsourcing.storage.StoreAggregate(aggregate)
}

func getAggregateEventsByMapping(oldAggregate *Aggregate, newAggregate *Aggregate) (events []IEvent, err error) {

	eventMapper := mapper.NewMapper(mapper.WithLogger(logger.Get()))

	var oldMappings map[string]interface{}
	var newMappings map[string]interface{}

	// old data
	if oldAggregate != nil {
		oldMappings, err = eventMapper.Map(reflect.ValueOf(oldAggregate.Data).Elem().Interface())
		if err != nil {
			logger.WithField("error", err.Error()).Error("error mapping aggregate")
			return nil, err
		}
	}

	// new data
	newMappings, err = eventMapper.Map(reflect.ValueOf(newAggregate.Data).Elem().Interface())
	if err != nil {
		logger.WithField("error", err.Error()).Error("error mapping aggregate")
		return nil, err
	}

	// validate the new object with the old
	for key, value := range newMappings {
		if _, ok := oldMappings[key]; ok {
			if oldMappings[key] != value {
				events = append(events, &Event{Id: common.NewULID(), Name: "updated_" + strings.ToLower(key), Data: map[string]interface{}{strings.ToLower(key): value}})
			}
		} else {
			events = append(events, &Event{Id: common.NewULID(), Name: "added_" + strings.ToLower(key), Data: map[string]interface{}{strings.ToLower(key): value}})
		}
	}

	// validate the data that was on the old object and it isn't on the new one
	for key, value := range oldMappings {
		if _, ok := newMappings[key]; !ok {
			events = append(events, &Event{Id: common.NewULID(), Name: "deleted_" + strings.ToLower(key), Data: map[string]interface{}{strings.ToLower(key): value}})
		}
	}

	return events, nil
}
