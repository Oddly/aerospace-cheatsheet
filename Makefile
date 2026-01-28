APP_NAME := cheatsheet
APP_ID := io.github.oddly.aerospace-cheatsheet
INSTALL_DIR := $(HOME)/.config/aerospace/bin

.PHONY: build build-darwin-arm64 build-darwin-amd64 build-linux-amd64 all clean install package-darwin package-linux

build:
	go build -o $(APP_NAME) .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o $(APP_NAME)-darwin-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o $(APP_NAME)-darwin-amd64 .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o $(APP_NAME)-linux-amd64 .

all: build-darwin-arm64 build-darwin-amd64 build-linux-amd64

clean:
	rm -f $(APP_NAME) $(APP_NAME)-darwin-* $(APP_NAME)-linux-*
	rm -rf "Aerospace Cheatsheet.app"

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(APP_NAME) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(APP_NAME)"

# Create proper macOS .app bundle with correct bundle ID
# Requires: go install fyne.io/fyne/v2/cmd/fyne@latest
package-darwin:
	fyne package --os darwin --name "Aerospace Cheatsheet" --appID $(APP_ID)

# Create Linux package
package-linux:
	fyne package --os linux --name "Aerospace Cheatsheet" --appID $(APP_ID)

# Install macOS app to /Applications
install-darwin: package-darwin
	rm -rf "/Applications/Aerospace Cheatsheet.app"
	cp -r "Aerospace Cheatsheet.app" /Applications/
	@echo "Installed to /Applications/Aerospace Cheatsheet.app"

# Cross-compilation using fyne-cross (Docker-based, recommended for cross-platform builds)
# Install: go install github.com/fyne-io/fyne-cross@latest
cross-darwin-arm64:
	fyne-cross darwin --arch arm64 --app-id $(APP_ID) --output $(APP_NAME)

cross-darwin-amd64:
	fyne-cross darwin --arch amd64 --app-id $(APP_ID) --output $(APP_NAME)

cross-linux-amd64:
	fyne-cross linux --arch amd64 --app-id $(APP_ID) --output $(APP_NAME)

cross-all: cross-darwin-arm64 cross-darwin-amd64 cross-linux-amd64
