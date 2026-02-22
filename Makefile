BINARY = bontree
INSTALL_DIR = /usr/local/bin
VERSION ?= dev

.PHONY: build install clean

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY) .

install: build
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)
