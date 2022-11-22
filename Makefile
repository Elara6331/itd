PREFIX ?= /usr/local
BIN_PREFIX = $(DESTDIR)$(PREFIX)/bin
SERVICE_PREFIX = $(DESTDIR)$(PREFIX)/lib/systemd/user
CFG_PREFIX = $(DESTDIR)/etc

all: version.txt
	go build
	go build ./cmd/itctl

clean:
	rm -f itctl
	rm -f itd
	rm -f version.txt

install:
	install -Dm755 ./itd $(BIN_PREFIX)/itd
	install -Dm755 ./itctl $(BIN_PREFIX)/itctl
	install -Dm644 ./itd.service $(SERVICE_PREFIX)/itd.service
	install -Dm644 ./itd.toml $(CFG_PREFIX)/itd.toml

uninstall:
	rm $(BIN_PREFIX)/itd
	rm $(BIN_PREFIX)/itctl
	rm $(SERVICE_PREFIX)/itd.service
	rm $(CFG_PREFIX)/itd.toml

version.txt:
	go generate

.PHONY: all clean install uninstall