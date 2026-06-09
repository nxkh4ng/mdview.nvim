BINARY_NAME = mdview-server
LOCAL_BUILD_PATH = $(HOME)/.local/share/nvim/mdview.nvim
GO_DIR = server

LDFLAGS = -s -w
GO_FLAGS = -trimpath -ldflags="$(LDFLAGS)"

build:
	CGO_ENABLED=0 go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/$(BINARY_NAME) ./...
	ls -lha $(LOCAL_BUILD_PATH)

run:
	$(LOCAL_BUILD_PATH)/$(BINARY_NAME)

clean:
	rm -rf $(LOCAL_BUILD_PATH)/$(BINARY_NAME)
	ls -lha $(LOCAL_BUILD_PATH)
