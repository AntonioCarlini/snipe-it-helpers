GO := go
BIN_DIR := bin

PROGRAM := $(BIN_DIR)/box-cat-assets-to-snipe-it
CATALOG := $(BIN_DIR)/asserts-import.csv
OUTPUT := $(BIN_DIR)/test.csv

.PHONY: all test testv clean

all: $(OUTPUT)

default: $(OUTPUT)

$(BIN_DIR):
	mkdir -p $@

# Build the Go executable.
$(PROGRAM): go.mod box-cat-assets-to-snipe-it/box-cat-assets-to-snipe-it.go | $(BIN_DIR)
	$(GO) build -o $@ box-cat-assets-to-snipe-it

$(BIN_DIR)/test.csv: $(CATALOG) $(PROGRAM) | $(BIN_DIR)
	$(PROGRAM) $(CATALOG) $(OUTPUT)

# Run Go tests.
test:
	$(GO) test ./...

testv:
	$(GO) test -v ./...

clean:
	rm -$(OUTPUT) $(PROGRAM)
