# Overridable version number
VERSION?=dev

# inject the version number into the Version variable
flags=-X 'github.com/inveracity/voki/internal/version.Version=$(VERSION)'

PHONY: build
build:
	@echo "Building..."
	@go mod tidy
	@CGO_ENABLED=0 go build -ldflags "$(flags)" -o bin/voki cmd/voki/main.go

PHONY: install
install: build
	cp bin/voki /home/$(USER)/.local/bin/voki

.PHONY: zip
zip: build
	@echo "Zipping..."
	@mkdir -p dist
	@zip -j dist/voki_linux_amd64.zip bin/voki
	@rm bin/voki

.PHONY: vault
vault:
	@echo "running vault..."
	VAULT_DEV_ROOT_TOKEN_ID=123456 vault server -dev
