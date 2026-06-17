BINARY_NAME = mdview
GO_DIR = server

LDFLAGS = -s -w
GO_FLAGS = -trimpath -ldflags="$(LDFLAGS)"

export CGO_ENABLED=0

dev:
	go build -C $(GO_DIR) $(GO_FLAGS) -o $(BINARY_NAME) ./...

run:
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
