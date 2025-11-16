APP_NAME := smartbash
VERSION ?= 1.0.0
ARCH ?= amd64

PKG_DIR := packaging
DEBIAN_DIR := $(PKG_DIR)/DEBIAN
BIN_DIR := $(PKG_DIR)/usr/bin

OUT_DIR := dist
BIN_OUT := $(OUT_DIR)/$(APP_NAME)_$(VERSION)_$(ARCH)
DEB_FILE := $(OUT_DIR)/$(APP_NAME)_$(VERSION)_$(ARCH).deb

.PHONY: build deb dirs clean build-all deb-all

# only binary
build:
	@echo "==> Building binary..."
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=$(ARCH) go build -o $(BIN_OUT) smartbash.go
	@chmod 755 $(BIN_OUT)
	@echo "Binary: $(BIN_OUT)"

# DEB dirs
dirs:
	@mkdir -p $(DEBIAN_DIR)
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(OUT_DIR)

# Create DEBIAN/control
$(DEBIAN_DIR)/control:
	@echo "Package: $(APP_NAME)"                   >  $(DEBIAN_DIR)/control
	@echo "Version: $(VERSION)"                   >> $(DEBIAN_DIR)/control
	@echo "Section: utils"                        >> $(DEBIAN_DIR)/control
	@echo "Priority: optional"                    >> $(DEBIAN_DIR)/control
	@echo "Architecture: $(ARCH)"                 >> $(DEBIAN_DIR)/control
	@echo "Maintainer: vladyka970@gmail.com"      >> $(DEBIAN_DIR)/control
	@echo "Description: Smart Bash fuzzy utility" >> $(DEBIAN_DIR)/control
	@echo " A fuzzy-history enhanced shell."      >> $(DEBIAN_DIR)/control

# Build DEB
deb: dirs $(DEBIAN_DIR)/control
	@echo "==> Building DEB package..."
	GOOS=linux GOARCH=$(ARCH) go build -o $(BIN_DIR)/$(APP_NAME) smartbash.go
	chmod 755 $(BIN_DIR)/$(APP_NAME)
	dpkg-deb --build $(PKG_DIR) $(DEB_FILE)
	@echo "DEB: $(DEB_FILE)"

# Building binary
build-all:
	$(MAKE) build ARCH=amd64

# Building DEB
deb-all:
	$(MAKE) deb ARCH=amd64

clean:
	@echo "Cleaning..."
	rm -rf dist
	rm -f $(BIN_DIR)/$(APP_NAME)
	rm -f $(DEBIAN_DIR)/control
