package models

import (
	"event-sourcing/common"
	"reflect"
	"strings"

	"github.com/joaosoft/logger"
	"github.com/joaosoft/manager"
	"github.com/joaosoft/mapper"
	"github.com/joaosoft/migration/services"
)

type IStorage interface {
	GetAggregate(id, typ string, obj interface{}) (aggregate *Aggregate, err error)
	StoreAggregate(aggregate *Aggregate) (err error)
}

type EventSourcing struct {
	storage       IStorage
	config        *EventSourcingConfig
	isLogExternal bool
	pm            *manager.Manager
	logger        logger.ILogger
}

func NewEventSourcing(options ...EventSourcingOption) (*EventSourcing, error) {
	config, simpleConfig, err := NewConfig()

	service := &EventSourcing{
		pm:     manager.NewManager(manager.WithRunInBackground(true)),
		logger: logger.NewLogDefault("event-sourcing", logger.WarnLevel),
		config: config.EventSourcing,
	}

	if service.isLogExternal {
		service.pm.Reconfigure(manager.WithLogger(logger.Instance))
	}

	if err != nil {
		service.logger.Error(err.Error())
	} else if config.EventSourcing != nil {
		service.pm.AddConfig("config_app", simpleConfig)
		level, _ := logger.ParseLevel(config.EventSourcing.Log.Level)
		service.logger.Debugf("setting log level to %s", level)
		service.logger.Reconfigure(logger.WithLevel(level))
	}

	service.Reconfigure(options...)

	// execute migrations
	migration, err := services.NewCmdService(services.WithCmdConfiguration(&service.config.Migration))
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if _, err := migration.Execute(services.OptionUp, 0, services.ExecutorModeDatabase); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	// database
	simpleDB := service.pm.NewSimpleDB(&config.EventSourcing.Db)
	if err := service.pm.AddDB("db_postgres", simpleDB); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if err = simpleDB.Start(); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	service.storage = NewStorage(simpleDB.Get(), service.logger)

	// initialize services
	if err := service.pm.Start(); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return service, nil

}
func (es *EventSourcing) Save(aggregate *Aggregate) (err error) {

	if len(aggregate.Events) == 0 {
		obj := reflect.New(reflect.TypeOf(aggregate.Data).Elem()).Interface()
		oldAggregate, err := es.storage.GetAggregate(aggregate.Id, aggregate.Type, obj)
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

	return es.storage.StoreAggregate(aggregate)
}

func getAggregateEventsByMapping(oldAggregate *Aggregate, newAggregate *Aggregate) (events []IEvent, err error) {
	eventMapper := mapper.NewMapper(mapper.WithLogger(logger.Instance))

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
