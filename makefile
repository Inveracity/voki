bin/voki:
	go build -o bin/voki cmd/voki/main.go

PHONY: build
build: bin/voki

PHONY: install
install: build
	cp bin/voki /home/$(USER)/.local/bin/voki
