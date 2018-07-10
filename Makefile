run:
	go run ./main.go

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