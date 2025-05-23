PACKAGE_NAME        	:= "github.com/safetyculture/safetyculture-exporter"
GOLANG_CROSS_VERSION  := v1.18.1

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: integration-tests
integration-tests:
	TEST_DB_DIALECT="postgres" TEST_DB_CONN_STRING="postgresql://safetyculture_exporter:safetyculture_exporter@localhost:5434/safetyculture_exporter_db" go test ./... -tags=sql
	TEST_DB_DIALECT="mysql" TEST_DB_CONN_STRING="root:safetyculture_exporter@tcp(localhost:3308)/safetyculture_exporter_db?charset=utf8mb4&parseTime=True&loc=Local" go test ./... -tags=sql
	TEST_DB_DIALECT="sqlserver" TEST_DB_CONN_STRING="sqlserver://sa:SafetyCultureExporter12345@localhost:1433?database=master" go test ./... -tags=sql

.PHONY: soak-tests
soak-tests:
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="postgres" TEST_DB_CONN_STRING="postgresql://safetyculture_exporter:safetyculture_exporter@localhost:5434/safetyculture_exporter_db" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="mysql" TEST_DB_CONN_STRING="root:safetyculture_exporter@tcp(localhost:3308)/safetyculture_exporter_db?charset=utf8mb4&parseTime=True&loc=Local" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="sqlserver" TEST_DB_CONN_STRING="sqlserver://sa:SafetyCultureExporter12345@localhost:1433?database=master" go test ./... -tags=soak
	TEST_API_HOST="https://api.safetyculture.io" TEST_DB_DIALECT="sqlite" TEST_DB_CONN_STRING="file::memory:" go test ./... -tags=soak

.PHONY: start-local-mssql
start-local-mssql:
	 docker-compose -f docker-compose-local-volume.yml up sqlserver

.PHONY: start-local-postgres
start-local-postgres:
	 docker-compose -f docker-compose-local-volume.yml up postgres
