BINARY   := guardrail

.PHONY: build test lint clean run

build:
	go build -o bin/$(BINARY) ./cmd/guardrail

run: build
	./bin/$(BINARY) watch

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -f bin/$(BINARY)
