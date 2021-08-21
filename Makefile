PREFIX ?= /usr/local
BIN_PREFIX = $(DESTDIR)$(PREFIX)/bin
SERVICE_PREFIX = $(DESTDIR)$(PREFIX)/lib/systemd/user
CFG_PREFIX = $(DESTDIR)/etc

all:
	go build
	go build ./cmd/itctl

install:
	install -Dm755 ./itd $(BIN_PREFIX)/itd
	install -Dm755 ./itctl $(BIN_PREFIX)/itctl
	install -Dm644 ./itd.service $(SERVICE_PREFIX)/itd.service
	install -Dm644 ./itd.toml $(CFG_PREFIX)/itd.toml

clean:
	rm -f itctl
	rm -f itd

.PHONY: all install clean