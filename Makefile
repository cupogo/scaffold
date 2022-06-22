.SILENT :
.PHONY: docs

NAME:=scaffold
ROOF:=hyyl.xyz/cupola/scaffold
DATE := $(shell date '+%Y%m%d')
TAG:=$(shell git describe --tags --always)
GO=$(shell which go)
GOMOD=$(shell echo "$${GO111MODULE:-auto}")


MDs=account.md

help:
	echo "make modcodegen"

docs:
	$(info docs: $(MDs))

generate:
	GO111MODULE=$(GOMOD) $(GO) generate ./...

modcodegen:
	mkdir -p ./pkg/models
	for name in $(MDs); do \
		echo $${name}; \
		GO111MODULE=$(GOMOD) $(GO) run -tags=sdkcodegen ./scripts/sdkcodegen docs/$${name} ./pkg/models/$${name}.go ; \
	done

vet:
	echo "Checking ./pkg/... ./cmd/... , with GOMOD=$(GOMOD)"
	GO111MODULE=$(GOMOD) $(GO) vet -all ./pkg/...

deps:
	GO111MODULE=on $(GO) install github.com/ddollar/forego@latest
	GO111MODULE=on $(GO) install github.com/liut/rerun@latest
	GO111MODULE=on $(GO) install github.com/swaggo/swag/cmd/swag@latest
	GO111MODULE=on $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint:
	GO111MODULE=$(GOMOD) golangci-lint --disable structcheck run ./pkg/...

clean:
	echo "Cleaning dist"
	rm -rf dist
	rm -f ./$(NAME)-*

showver:
	echo "version: $(TAG)"

dist/linux_amd64/$(NAME): $(SOURCES) showver docs/swagger.yaml
	echo "Building $(NAME) of linux"
	mkdir -p dist/linux_amd64 && GO111MODULE=$(GOMOD) GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -s -w" -o dist/linux_amd64/$(NAME) .

dist/darwin_amd64/$(NAME): $(SOURCES) showver docs/swagger.yaml
	echo "Building $(NAME) of darwin"
	mkdir -p dist/darwin_amd64 && GO111MODULE=$(GOMOD) GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS) -w" -o dist/darwin_amd64/$(NAME) .

dist: vet dist/linux_amd64/$(NAME) dist/darwin_amd64/$(NAME)

package: dist
	echo "Packaging $(NAME)"
	ls dist/linux_amd64 | xargs tar -cvJf $(NAME)-linux-amd64-$(TAG).tar.xz -C dist/linux_amd64
	ls dist/darwin_amd64 | xargs tar -cvJf $(NAME)-darwin-amd64-$(TAG).tar.xz -C dist/darwin_amd64
