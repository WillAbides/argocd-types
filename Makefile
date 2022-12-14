GOCMD=go
GOBUILD=$(GOCMD) build
PATH := "${CURDIR}/bin:$(PATH)"

.PHONY: gobuildcache

bin/golangci-lint:
	script/bindown install $(notdir $@)

bin/shellcheck:
	script/bindown install $(notdir $@)

bin/gofumpt:
	script/bindown install $(notdir $@)

HANDCRAFTED_REV := 082e94edadf89c33db0afb48889c8419a2cb46a9
bin/handcrafted:
	GOBIN=${CURDIR}/bin \
	go install github.com/willabides/handcrafted@$(HANDCRAFTED_REV)

GOIMPORTS_REV := v0.4.0
bin/goimports:
	GOBIN=${CURDIR}/bin \
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_REV)
