# Makefile for Go project automation

.PHONY: all build test lint fmt vet coverage clean check-all

all: build test

build:
	@echo "> Building Go project..."
	go build ./...

test:
	@echo "> Running all Go tests..."
	go test ./... -v

lint:
	@echo "> Running golangci-lint..."
	golangci-lint run ./...

fmt:
	@echo "> Checking code formatting..."
	gofmt -l -s .

fmt-fix:
	@echo "> Formatting code..."
	gofmt -w -s .

vet:
	@echo "> Running go vet..."
	go vet ./...

coverage:
	@echo "> Running test coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	@echo "> Cleaning build artifacts..."
	rm -f coverage.out

check-all: fmt vet lint build test coverage
	@echo "> All checks complete."

# To use golangci-lint, install it with:
# go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
