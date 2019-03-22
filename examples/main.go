package main

import (
	"event-sourcing/common"
	"event-sourcing/models"
	"fmt"
	"time"

	"github.com/joaosoft/logger"
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
	eventSourcing, err := models.NewEventSourcing()
	if err != nil {
		panic(err)
	}

	// person - with managed events by the user
	aggregate1 := models.NewAggregate("persons", "person", &Person{
		Name: "joao",
		Age:  30,
	})
	aggregate1.AddEventHandler("person_change_name", handler_person_change_name)
	aggregate1.Causes(&models.Event{
		Id:        common.NewULID(),
		Name:      "person_changed_name",
		Data:      PersonChangeNameEvent{Name: "manuel"},
		CreatedAt: time.Now(),
	})
	err = eventSourcing.Save(aggregate1)
	if err != nil {
		fmt.Println(err)
	}

	// address - with automatic event generation
	aggregate2 := models.NewAggregate("addresses", "address", &Address{
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
