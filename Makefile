PROJECTNAME := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)(\.git)?$$|\1|' || basename "$$(pwd)")
PROJECTORG  := $(shell git remote get-url origin 2>/dev/null | sed -E 's|.*/([^/]+)/[^/]+(\.git)?$$|\1|' || basename "$$(dirname "$$(pwd)")")

VERSION    ?= $(shell cat release.txt 2>/dev/null || echo "devel")
BUILD_DATE := $(shell date +"%a %b %d, %Y at %H:%M:%S %Z")
COMMIT_ID  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "N/A")
PLATFORMS  ?= linux/amd64,linux/arm64

OFFICIAL_SITE ?= https://github.com/$(PROJECTORG)/$(PROJECTNAME)

LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.CommitID=$(COMMIT_ID)' \
	-X 'main.BuildDate=$(BUILD_DATE)' \
	-X 'main.OfficialSite=$(OFFICIAL_SITE)'

BINDIR   := binaries
RELDIR   := releases
REGISTRY := ghcr.io/$(PROJECTORG)/$(PROJECTNAME)

GO_VOL ?= go-state

GO_DOCKER := docker run --rm -it \
	--name $(PROJECTNAME)-$$(tr -dc 'a-z0-9' </dev/urandom | head -c8) \
	-v $(PWD):/app \
	-v $(GO_VOL):/usr/local/share/go \
	-w /app \
	casjaysdev/go:latest

.PHONY: help build release docker test dev lint clean

help: ## Show this help message
	@printf '\n\033[1;37m  %s v%s\033[0m\n\n' "$(PROJECTNAME)" "$(VERSION)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-38s\033[0m- %s\n", $$1, $$2}'
	@printf '\n'

build: ## Build binaries for all 8 platforms
	@mkdir -p $(BINDIR)
	$(GO_DOCKER) sh -c '\
		CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-linux-amd64 ./src && \
		CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-linux-arm64 ./src && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-darwin-amd64 ./src && \
		CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-darwin-arm64 ./src && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-windows-amd64.exe ./src && \
		CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-windows-arm64.exe ./src && \
		CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-freebsd-amd64 ./src && \
		CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINDIR)/$(PROJECTNAME)-freebsd-arm64 ./src'

dev: ## Quick single-platform build into a temp dir
	@mkdir -p "$${TMPDIR:-/tmp}/$(PROJECTORG)"
	@BUILD_DIR=$$(mktemp -d "$${TMPDIR:-/tmp}/$(PROJECTORG)/$(PROJECTNAME)-XXXXXX") && \
		echo "Building dev binary..." && \
		docker run --rm -it \
			--name $(PROJECTNAME)-$$(tr -dc 'a-z0-9' </dev/urandom | head -c8) \
			-v $(PWD):/app \
			-v $(GO_VOL):/usr/local/share/go \
			-w /app \
			casjaysdev/go:latest \
			sh -c 'CGO_ENABLED=0 go build -o $$BUILD_DIR/$(PROJECTNAME) ./src' && \
		echo "Built: $$BUILD_DIR/$(PROJECTNAME)"

test: ## Run vet and tests
	$(GO_DOCKER) sh -c 'CGO_ENABLED=0 go vet ./... && CGO_ENABLED=0 go test -v -cover ./...'

lint: ## Run linter
	$(GO_DOCKER) sh -c 'CGO_ENABLED=0 golangci-lint run ./...'

docker: ## Build multi-arch Docker image locally
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--build-arg BUILD_DATE="$(BUILD_DATE)" \
		-t $(REGISTRY):$(VERSION) \
		-t $(REGISTRY):latest \
		-f docker/Dockerfile .

release: ## Prepare release archives
	@mkdir -p $(RELDIR)
	@$(MAKE) build
	@for f in $(BINDIR)/$(PROJECTNAME)-*; do \
		name=$$(basename $$f); \
		tar -czf $(RELDIR)/$$name.tar.gz -C $(BINDIR) $$name; \
	done
	@echo "Release archives in $(RELDIR)/"

clean: ## Remove build artifacts
	@rm -rf $(BINDIR) $(RELDIR)
