# Variables
APP_NAME := form3-interview-simulator
SRC_DIR := ./
BUILD_DIR := ./bin
GO := /home/lucas/go/go1.23.4/bin/go
TEST_FLAGS := -v -count=1
LDFLAGS := -s -w
GOFILES := $(shell find . -name '*.go' -not -path "./vendor/*")

# Default target
.PHONY: all
all: build

# Build the project
.PHONY: build
build: $(BUILD_DIR)/$(APP_NAME)

$(BUILD_DIR)/$(APP_NAME): $(GOFILES)
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)

# Run the project
.PHONY: run
run: build
	@echo "Running the project..."
	$(BUILD_DIR)/$(APP_NAME)

# Test the project
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./... $(TEST_FLAGS)

# Clean the build directory
.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Run everything (build, test, run)
.PHONY: all
all: test build run
