.PHONY: swagger test integration-test build run

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d ./cmd/app,./internal -o docs --parseInternal

# Run unit tests
test:
	@echo "Running unit tests..."
	@go test -v -coverprofile=coverage.out ./...

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./integration/...

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/fitslot ./cmd/app

# Run the application
run:
	@go run ./cmd/app

# Install dependencies
deps:
	@go mod download
	@go mod tidy

# Install Swagger CLI
install-swagger:
	@go install github.com/swaggo/swag/cmd/swag@v1.16.6


