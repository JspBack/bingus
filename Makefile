BINARY_NAME=bingus.exe
BUILD_DIR=build
GO=go
GOFLAGS=-ldflags="-s -w" -trimpath
APP_FOLDER=./cobra/cmd/bingus

.PHONY: all
all: build run

.PHONY: build
build:
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(APP_FOLDER)

.PHONY: build-all
build-all:
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/bingus-cbr.exe ./cobra/cmd/bingus
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/bingus-bta.exe ./bta/cmd/bingus

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)

.PHONY: run
run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)