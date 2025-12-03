.PHONY: build-backend build-frontend build clean test dev

# Build backend Go plugin
build-backend:
	@echo "Building backend..."
	@go mod tidy
	@mkdir -p dist
	@go build -o dist/gpx_grafana-connect .

# Build frontend
build-frontend:
	@echo "Building frontend..."
	@npm install
	@npm run build

# Build everything
build: build-backend build-frontend
	@echo "Build complete!"

# Development mode
dev:
	@echo "Starting development mode..."
	@npm run dev

# Run tests
test:
	@echo "Running tests..."
	@go test ./...
	@npm run test

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf dist/
	@rm -rf node_modules/
	@rm -rf .grafana/
	@go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@npm install

