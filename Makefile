.PHONY: build test lint run clean

VERSION ?= dev

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/zvault ./cmd/zvault

test:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run ./cmd/zvault

clean:
	rm -rf bin/
