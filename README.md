# event-sourcing
[![Build Status](https://travis-ci.org/joaosoft/event-sourcing.svg?branch=master)](https://travis-ci.org/joaosoft/event-sourcing) | [![codecov](https://codecov.io/gh/joaosoft/event-sourcing/branch/master/graph/badge.svg)](https://codecov.io/gh/joaosoft/event-sourcing) | [![Go Report Card](https://goreportcard.com/badge/github.com/joaosoft/event-sourcing)](https://goreportcard.com/report/github.com/joaosoft/event-sourcing) | [![GoDoc](https://godoc.org/github.com/joaosoft/event-sourcing?status.svg)](https://godoc.org/github.com/joaosoft/event-sourcing)

A simplified event-sourcing that allows you to add complexity depending of your requirements.
The easy way to use the event-sourcing:

## Installation
```
docker-compose up -d postgres

make init
make migrate
make run
```

## Usage
```go
import log github.com/joaosoft/event-sourcing

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
    Number: 10,
})
err = eventSourcing.Save(aggregate2)
if err != nil {
    fmt.Println(err)
}

fmt.Println("DONE")
```

###### If i miss something or you have something interesting, please be part of this project. Let me know! My contact is at the end.

## With support for
* add manual events
* auto generate events
  
## Dependecy Management 
>### Dep

Project dependencies are managed using Dep. Read more about [Dep](https://github.com/golang/dep).
* Install dependencies: `dep ensure`
* Update dependencies: `dep ensure -update`


>### Go
```
go get github.com/joaosoft/event-sourcing
```

## Usage 
This examples are available in the project at [event-sourcing/main.go](https://github.com/joaosoft/event-sourcing/tree/master/main.go)

## Known issues
* all the maps do not guarantee order of the items! 


## Follow me at
Facebook: https://www.facebook.com/joaosoft

LinkedIn: https://www.linkedin.com/in/jo%C3%A3o-ribeiro-b2775438/

##### If you have something to add, please let me know joaosoft@gmail.com
