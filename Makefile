PROTOC = protoc
PROTOC_GEN_GO = protoc-gen-go
PROTOC_GEN_GRPC_GO = protoc-gen-go-grpc

# Directories
CONTRACT_DIR = contract
SRC_DIR = src
# Proto files
PROTO_FILES = $(wildcard $(CONTRACT_DIR)/*.proto)

# Targets
all: generate

generate:
	@for proto in $(PROTO_FILES); do \
		base_name=$$(basename $$proto .proto); \
		mkdir -p $(SRC_DIR)/$$base_name; \
		$(PROTOC) --go_out=$(SRC_DIR)/$$base_name --go_opt=paths=source_relative \
		--go-grpc_out=$(SRC_DIR)/$$base_name --go-grpc_opt=paths=source_relative \
		$$proto; \
		mv $(SRC_DIR)/$$base_name/$(CONTRACT_DIR)/* $(SRC_DIR)/$$base_name; \
		rmdir $(SRC_DIR)/$$base_name/$(CONTRACT_DIR); \
	done

clean:
	rm -f $(SRC_DIR)/*/*.pb.go

