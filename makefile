PHONY: build
build:
	go build -o bin/voki cmd/voki/main.go

PHONY: install
install: build
	cp bin/voki /home/$(USER)/.local/bin/voki
