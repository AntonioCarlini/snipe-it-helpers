GO := go
BIN_DIR := bin
DATA_DIR := data

PROGRAM := $(BIN_DIR)/box-cat-assets-to-snipe-it
CATALOG := $(DATA_DIR)/asserts-import.csv
OUTPUT := $(BIN_DIR)/test.csv

.PHONY: all test testv clean

all: $(OUTPUT)

default: $(OUTPUT)

$(BIN_DIR):
	mkdir -p $@

# Build the Go executable.
$(PROGRAM): go.mod ./box-cat-assets-to-snipe-it | $(BIN_DIR)
	$(GO) build -o $@ ./box-cat-assets-to-snipe-it

$(OUTPUT): $(CATALOG) $(PROGRAM) | $(BIN_DIR)
	$(PROGRAM) $(CATALOG) $(OUTPUT)

# Run Go tests.
test:
	$(GO) test ./...

testv:
	$(GO) test -v ./...

clean:
	@rm -rf $(BIN_DIR)
