package models

import (
	"fmt"
	"github.com/joaosoft/migration/services"
	"github.com/joaosoft/manager"
)

// AppConfig ...
type AppConfig struct {
	EventSourcing *EventSourcingConfig `json:"event-sourcing"`
}

// UploaderConfig ...
type EventSourcingConfig struct {
	Db        manager.DBConfig         `json:"db"`
	Migration services.MigrationConfig `json:"migration"`
	Log       struct {
		Level string `json:"level"`
	} `json:"log"`
}

// NewConfig ...
func NewConfig() (*AppConfig, manager.IConfig, error) {
	appConfig := &AppConfig{}
	simpleConfig, err := manager.NewSimpleConfig(fmt.Sprintf("/config/app.%s.json", GetEnv()), appConfig)

	return appConfig, simpleConfig, err
}
