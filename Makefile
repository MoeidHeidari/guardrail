BINARY   := guardrail
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  := -ldflags "-X github.com/MoeidHeidari/guardrail/cmd.Version=$(VERSION) \
                       -X github.com/MoeidHeidari/guardrail/cmd.Commit=$(COMMIT) \
                       -X github.com/MoeidHeidari/guardrail/cmd.BuildDate=$(DATE)"

.PHONY: build test lint clean run

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

run: build
	./bin/$(BINARY) --help

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
