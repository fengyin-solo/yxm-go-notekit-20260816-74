.PHONY: build test test-race vet clean run

BINARY=notekit
GO=go
GOFLAGS=-v

build:
	$(GO) build $(GOFLAGS) -o $(BINARY) ./cmd/notekit

test:
	$(GO) test -count=1 ./...

test-race:
	$(GO) test -race -count=1 ./...

vet:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)
