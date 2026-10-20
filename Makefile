.PHONY: build test clean install run-test fmt lint

# Build the binary
build:
	go build -o shrike .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/shrike-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o dist/shrike-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/shrike-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/shrike-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/shrike-windows-amd64.exe .

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Run shrike on test repository
run-test: build
	./shrike scan test-repo

# Run shrike with verbose output
run-test-verbose: build
	./shrike scan test-repo --verbose

# Run shrike and output JSON
run-test-json: build
	./shrike scan test-repo --json

# Run shrike and output SARIF
run-test-sarif: build
	./shrike scan test-repo --sarif

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...
	golangci-lint run

# Clean build artifacts
clean:
	rm -f shrike
	rm -rf dist/
	rm -f shrike-findings.md
	rm -f shrike-findings.json
	rm -f shrike-findings.sarif

# Install shrike to $GOPATH/bin
install:
	go install .

# Show help
help:
	@echo "Shrike - AI agent in CI security scanner"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build shrike binary"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make test           - Run tests"
	@echo "  make run-test       - Run shrike on test repository"
	@echo "  make run-test-verbose - Run shrike with verbose output"
	@echo "  make run-test-json  - Output findings as JSON"
	@echo "  make run-test-sarif - Output findings as SARIF"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install to $$GOPATH/bin"
	@echo "  make help           - Show this help message"
