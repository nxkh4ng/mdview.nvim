BINARY_NAME = mdview
GO_DIR = server
LOCAL_BUILD_PATH = $(HOME)/.local/share/nvim/mdview.nvim

LDFLAGS = -s -w
GO_FLAGS = -trimpath -ldflags="$(LDFLAGS)"

export CGO_ENABLED=0

dev:
	go build -C $(GO_DIR) $(GO_FLAGS) -o $(LOCAL_BUILD_PATH)/$(BINARY_NAME) ./...
	ls -lha $(LOCAL_BUILD_PATH)

run:
	$(LOCAL_BUILD_PATH)/$(BINARY_NAME)
	ls -lha $(LOCAL_BUILD_PATH)

clean:
	rm -f $(LOCAL_BUILD_PATH)/$(BINARY_NAME)
	ls -lha $(LOCAL_BUILD_PATH)
