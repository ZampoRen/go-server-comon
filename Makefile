.PHONY: build run proto clean install-proto-tools tidy

# Install protobuf tools
install-proto-tools:
	@echo "Installing protobuf tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Done. Make sure protoc is installed: https://grpc.io/docs/protoc-installation/"

# Update dependencies
tidy:
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

# Generate proto files
proto:
	@echo "Generating proto files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/user/user.proto
	@echo "Proto files generated successfully"

# Build all services
build: proto
	@echo "Building services..."
	@go build -o bin/user ./cmd/user
	@go build -o bin/ws-bench ./cmd/ws-bench
	@go build -o bin/ws-heartbeat ./cmd/ws-heartbeat

# Build Windows version
build-windows: proto
	@echo "Building Windows version..."
	@GOOS=windows GOARCH=amd64 go build -o bin/user.exe ./cmd/user
	@GOOS=windows GOARCH=amd64 go build -o bin/ws-bench.exe ./cmd/ws-bench
	@GOOS=windows GOARCH=amd64 go build -o bin/ws-heartbeat.exe ./cmd/ws-heartbeat

# Run user service
run-user: proto
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
	@find . -name "*.pb.go" -delete
	@find . -name "*.pb.gw.go" -delete

