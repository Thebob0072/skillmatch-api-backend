# Makefile for SkillMatch API Backend

.PHONY: help test test-verbose test-coverage run build clean

# Default target
help:
	@echo "Available commands:"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make run            - Run the application"
	@echo "  make build          - Build the application"
	@echo "  make clean          - Clean build artifacts"

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for specific package
test-auth:
	@echo "Running auth tests..."
	go test -v -run TestAuth

test-booking:
	@echo "Running booking tests..."
	go test -v -run TestBooking

test-financial:
	@echo "Running financial tests..."
	go test -v -run TestFinancial

# Run the application
run:
	@echo "Starting application..."
	go run .

# Build the application
build:
	@echo "Building application..."
	go build -o bin/skillmatch-api .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Run all checks (test + lint + fmt)
check: fmt lint test
	@echo "All checks passed!"
