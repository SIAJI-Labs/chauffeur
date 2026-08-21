.PHONY: build panel-assets dev install clean test

BIN     := chauf
BUILD   := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

panel-assets:
	cd internal/panel-apps && npm run build
	rm -rf internal/panel/static
	mkdir -p internal/panel/static
	cp -R internal/panel-apps/dist/* internal/panel/static/

build: panel-assets
	go build $(LDFLAGS) -o $(BUILD)/$(BIN) ./cmd/chauf

dev:
	go run ./cmd/chauf $(ARGS)

install: build
	install -Dm755 $(BUILD)/$(BIN) ~/.chauffeur/bin/$(BIN)
	@echo "Installed to ~/.chauffeur/bin/chauf"

clean:
	rm -rf $(BUILD)/

test:
	go test ./...

# Run a quick sanity check
check: build
	./$(BUILD)/$(BIN) --version
	./$(BUILD)/$(BIN) --help
