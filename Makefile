init:
	go get github.com/joaosoft/migration

run:
	go run ./examples/main.go

build:
	go build ./...

fmt:
	go fmt ./...

vet:
	go vet ./*

meta:
	gometalinter ./*

migrate:
	migration up