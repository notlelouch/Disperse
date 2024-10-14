build:
	go build -o bin/server

run: build
	go run main.go

test:
	go test -v ./...

