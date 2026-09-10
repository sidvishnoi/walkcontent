BINARY := bin/walkcontent
PKG := .

.PHONY: all build run test vet fmt lint tools tidy clean

all: fmt vet test build

build:
	go build -trimpath -ldflags="-s -w" -o $(BINARY) $(PKG)

run: build
	$(BINARY)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	staticcheck ./...
	golangci-lint run ./...

tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
