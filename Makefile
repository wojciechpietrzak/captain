# Define variables
PROTOC = protoc
PROTOC_GEN_GO = $(shell which protoc-gen-go)
PROTOC_GEN_GO_GRPC = $(shell which protoc-gen-go-grpc)

PROTO_DIR = src/contract
OUT_DIR = .

PROTO_FILES = $(wildcard $(PROTO_DIR)/*.proto)

# Default target
all: check-plugins generate

# Check if required plugins are installed
check-plugins:
ifndef PROTOC_GEN_GO
	$(error "protoc-gen-go is not installed. Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
endif
ifndef PROTOC_GEN_GO_GRPC
	$(error "protoc-gen-go-grpc is not installed. Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
endif

# Generate Go code from .proto files
generate:
	$(PROTOC) -I src --go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) $(PROTO_FILES)

# Clean generated files (optional)
clean:
	rm -f src/pairing_engine/*.pb.go
