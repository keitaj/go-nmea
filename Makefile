.PHONY: test build run clean lint fmt vet

# Build the CLI tool
build:
	go build -o bin/nmea-cli ./cmd/nmea-cli

# Run tests
test:
	go test -v -race ./pkg/nmea/...

# Run go vet
vet:
	go vet ./...

# Run tests with coverage
cover:
	go test -cover -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	gofmt -w .

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Parse sample data
demo: build
	./bin/nmea-cli -f testdata/sample_yokohama.nmea -errors

# Parse only GGA sentences with verbose output
demo-gga: build
	./bin/nmea-cli -f testdata/sample_yokohama.nmea -type GGA -v

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
