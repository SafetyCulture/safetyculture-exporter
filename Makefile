PACKAGE_NAME        	:= "github.com/safetyculture/iauditor-exporter"
GOLANG_VERSION        ?= 1.15.3
GOLANG_CROSS_VERSION  := v$(GOLANG_VERSION)

.PHONY: docs
PANDOC ?= pandoc
%.1: %.1.md
	$(PANDOC) -s --to man -o $@ $<
%.5: %.5.md
	$(PANDOC) -s --to man -o $@ $<

docs: $(shell find ./docs/man -name "*.[1-7].md" | xargs echo | sed 's/\.md//g') ## Generates all the man pages

.PHONY: lint
lint: ## Runs more than 20 different linters using golangci-lint to ensure consistency in code.
ifeq ($(shell which golangci-lint),)
	brew install golangci/tap/golangci-lint
endif
	golangci-lint run

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: release-snapshot
release-snapshot:
	docker run \
		--rm \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		-f .goreleaser.yml --rm-dist --snapshot --skip-validate --skip-publish

.PHONY: release-dry-run
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		-f .goreleaser.yml --rm-dist --skip-validate --skip-publish

.PHONY: release
release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		troian/golang-cross:${GOLANG_CROSS_VERSION} \
		-f .goreleaser.yml release --rm-dist
