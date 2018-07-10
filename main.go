package main

import (
	"database/sql"
	"event-sourcing/common"
	"event-sourcing/models"
	"event-sourcing/storage"
	"fmt"
	"github.com/joaosoft/logger"
	"time"
)

func init() {
	logger.Reconfigure(logger.WithOptTag("service", "event-sourcing"))
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Address struct {
	Street string `json:"street"`
	Number int    `json:"number"`
}

func main() {
	conn, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	eventSourcing := models.NewEventSourcing(storage.NewStorage(conn))

	// person - with managed events by the user
	aggregate1 := models.NewAggregate("person_001", "person", &Person{
		Name: "joao",
		Age:  30,
	})
	aggregate1.AddEventHandler("person_change_name", handler_person_change_name)
	aggregate1.Causes(&models.Event{
		Id:        common.NewULID(),
		Name:      "person_change_name",
		Data:      PersonChangeNameEvent{Name: "manuel"},
		CreatedAt: time.Now(),
	})
	err = eventSourcing.Save(aggregate1)
	if err != nil {
		fmt.Println(err)
	}

	// address - with automatic event generation
	aggregate2 := models.NewAggregate("address_001", "address", &Address{
		Street: "caminho do senhor da luz",
		Number: 7,
	})
	err = eventSourcing.Save(aggregate2)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("DONE")
}

func handler_person_change_name(aggregate *models.Aggregate, event models.IEvent) error {
	aggregate.Data.(*Person).Name = event.GetData().(PersonChangeNameEvent).Name
	return nil
}

type PersonChangeNameEvent struct {
	Name string `json:"name"`
}
