build:
	go build -o ./build/output ./cmd/server

run-prod:	build
	./build/output

clean:
	rm -f ./build/output

run:
	go run ./cmd/server/main.go

test:
	go test -v ./...
