.PHONY: build test lint run clean

build:
	go build -o bin/zvault ./cmd/zvault

test:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run ./cmd/zvault

clean:
	rm -rf bin/
