# Go-Inject Makefile

.PHONY: help build test lint fmt vet clean coverage bench examples install-deps

# Default target
help: ## Show this help message
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the project
build: ## Build the project
	@echo "Building project..."
	go build ./...

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench: ## Run benchmark tests
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint code
lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Vet code
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean: ## Clean build artifacts and temporary files
	@echo "Cleaning..."
	go clean
	rm -f coverage.out coverage.html

# Install development dependencies
install-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run all checks (format, vet, lint, test)
check: fmt vet lint test ## Run all checks

# Run examples
examples: ## Run all examples
	@echo "Running basic example..."
	go run ./examples/basic/main.go
	@echo ""
	@echo "Running testing example..."
	go run ./examples/testing/main.go

# Continuous integration target
ci: check coverage ## Run CI pipeline (checks + coverage)

# Development setup
setup: install-deps ## Setup development environment
	@echo "Development environment ready!"

# Release preparation
prepare-release: clean check coverage ## Prepare for release
	@echo "Release preparation complete!"

# Initialize go module (run once)
init: ## Initialize go module
	go mod init github.com/go-inject/go-inject
	go mod tidy

# Update dependencies
update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Security check
security: ## Run security checks
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	gosec ./...

# Documentation generation
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Quick development cycle
dev: fmt vet test ## Quick development cycle (format, vet, test)