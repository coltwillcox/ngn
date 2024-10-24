BIN ?= ./bin
GO ?= go
GOFMT ?= gofmt
CP ?= sudo cp


$(BIN):
	mkdir -p $(BIN)

.PHONY: build
build: ./bin
	$(GO) build -o $(BIN) ./...

.PHONY: clean
clean:
	rm -rf $(BIN)/

.PHONY: test
test:
	$(GO) test -count=1 -race -v ./...

.PHONY: lint
lint:
	@$(GOFMT) -d -e .

.PHONY: install
install: build
	$(CP) ./bin/notilogd $(PREFIX)/usr/bin/
	$(CP) ./bin/notilogctl $(PREFIX)/usr/bin/
	$(CP) ./contrib/systemd/notilogd.service $(PREFIX)/usr/lib/systemd/user/
