.PHONY: build run clean install-tools tidy init-api update-api build-windows run-user run-ws-bench run-ws-heartbeat gen-grpc

# Install required tools
install-tools:
	@echo "Installing required tools..."
	@go install github.com/cloudwego/hertz/cmd/hz@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Done. Make sure protoc is installed: https://grpc.io/docs/protoc-installation/"

# Initialize Hertz API (first time only)
init-api:
	@echo "Initializing Hertz API..."
	@sh hz_gen.sh init

# Update Hertz API code (HTTP routes, handlers, models)
update-api:
	@echo "Updating Hertz API code..."
	@sh hz_gen.sh

# Generate gRPC code (if needed separately)
gen-grpc:
	@echo "Generating gRPC code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/user/user.proto
	@echo "gRPC code generated successfully"

# Update dependencies
tidy:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

# Build all services
build: update-api
	@echo "Building services..."
	@go build -o bin/user ./cmd/user
	@go build -o bin/ws-bench ./cmd/ws-bench
	@go build -o bin/ws-heartbeat ./cmd/ws-heartbeat

# Build Windows version
build-windows: update-api
	@echo "Building Windows version..."
	@GOOS=windows GOARCH=amd64 go build -o bin/user.exe ./cmd/user
	@GOOS=windows GOARCH=amd64 go build -o bin/ws-bench.exe ./cmd/ws-bench
	@GOOS=windows GOARCH=amd64 go build -o bin/ws-heartbeat.exe ./cmd/ws-heartbeat

# Run user service
run-user: update-api
	@echo "Running user service..."
	@go run ./cmd/user

# Run WebSocket benchmark
run-ws-bench:
	@echo "Running WebSocket benchmark..."
	@go run ./cmd/ws-bench

# Run WebSocket heartbeat test
run-ws-heartbeat:
	@echo "Running WebSocket heartbeat test..."
	@go run ./cmd/ws-heartbeat

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Note: Generated files in biz/ directory are preserved. Use 'make update-api' to regenerate."

