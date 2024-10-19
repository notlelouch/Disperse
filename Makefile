build:
	go build -o ./build/output ./cmd/server

run: build
	./build/output

test:
	go test -v ./...

