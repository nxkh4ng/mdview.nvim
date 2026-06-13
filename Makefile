BINARY_NAME = mdview
LOCAL_BUILD_PATH = $(HOME)/.local/share/nvim/mdview.nvim
GO_DIR = server

LDFLAGS = -s -w
GO_FLAGS = -trimpath -ldflags="$(LDFLAGS)"

# Detect OS and arch cho target dev
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Map OS
ifeq ($(UNAME_S),Linux)
	OS_SUFFIX := linux
else ifeq ($(UNAME_S),Darwin)
	OS_SUFFIX := darwin
else
	OS_SUFFIX := windows
endif

# Map arch
ifeq ($(UNAME_M),x86_64)
	ARCH_SUFFIX := amd64
else ifeq ($(UNAME_M),aarch64)
	ARCH_SUFFIX := arm64
endif

dev:
	CGO_ENABLED=0 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/$(BINARY_NAME)-$(OS_SUFFIX)-$(ARCH_SUFFIX) ./...
	ls -lha $(LOCAL_BUILD_PATH)

run:
	$(LOCAL_BUILD_PATH)/$(BINARY_NAME)

clean:
	rm -rf $(LOCAL_BUILD_PATH)/$(BINARY_NAME)*
	ls -lha $(LOCAL_BUILD_PATH)

build-all:
	@echo "Building for linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-linux-amd64 ./...
	@echo "Building for linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-linux-arm64 ./...
	@echo "Building for darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-darwin-amd64 ./...
	@echo "Building for darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-darwin-arm64 ./...
	@echo "Building for windows/amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-windows-amd64.exe ./...
	@echo "Building for windows/arm64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/mdview-windows-arm64.exe ./...
	@echo "Done."
	@ls -lha $(LOCAL_BUILD_PATH)
