PACKAGE_NAME        	:= "github.com/safetyculture/iauditor-exporter"
GOLANG_VERSION        ?= 1.15.3
GOLANG_CROSS_VERSION  := v$(GOLANG_VERSION)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: integration-tests
integration-tests:
	TEST_DB_DIALECT="postgres" TEST_DB_CONN_STRING="postgresql://iauditor_exporter:iauditor_exporter@localhost:5434/iauditor_exporter_db" go test ./... -tags=sql
	TEST_DB_DIALECT="mysql" TEST_DB_CONN_STRING="root:iauditor_exporter@tcp(localhost:3308)/iauditor_exporter_db?charset=utf8mb4&parseTime=True&loc=Local" go test ./... -tags=sql
	TEST_DB_DIALECT="sqlserver" TEST_DB_CONN_STRING="sqlserver://sa:iAuditorExporter12345@localhost:1433?database=master" go test ./... -tags=sql

.PHONY: soak-tests
soak-tests:
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="postgres" TEST_DB_CONN_STRING="postgresql://iauditor_exporter:iauditor_exporter@localhost:5434/iauditor_exporter_db" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="mysql" TEST_DB_CONN_STRING="root:iauditor_exporter@tcp(localhost:3308)/iauditor_exporter_db?charset=utf8mb4&parseTime=True&loc=Local" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="sqlserver" TEST_DB_CONN_STRING="sqlserver://sa:iAuditorExporter12345@localhost:1433?database=master" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="sqlite" TEST_DB_CONN_STRING="file::memory:" go test ./... -tags=soak

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
