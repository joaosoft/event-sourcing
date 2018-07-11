init:
	go get github.com/rubenv/sql-migrate/...

run:
	go run ./examples/main.go

build:
	go build .

fmt:
	go fmt ./...

vet:
	go vet ./*

meta:
	gometalinter ./*

migrate:
	sql-migrate up --config=config/dbconfig.yml