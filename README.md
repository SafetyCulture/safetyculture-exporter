# iAuditor Exporter

[![Maintainability](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/maintainability)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/test_coverage)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/test_coverage)

iAuditor Exporter is a CLI tool for extracting your iAuditor data.

To see the full set of commands available run `iauditor-exporter --help`

## Install

Download the latest release from [iauditor-exporter/releases](https://github.com/SafetyCulture/iauditor-exporter/releases).

### Quick Start

To get up and running quickly run the `configure` command to create a stub config file.

`./iauditor-exporter configure`

This will create a config file `.iauditor-exporter.yaml` in the same directory. Use `--config-path "/my/path/.iauditor-exporter.yaml"` to set a custom path.

Modify this file adding an `access-token` and modifying the `export` options to customise the data exported by the tool.

Run `./iauditor-exporter csv` to create a CSV export of your data.

## Configure

`iauditor-exporter configure` can be used to initialise your configuration file. Use `iauditor-exporter configure --help` to see the options available.

An example configuration file looks like this.

```yaml
access_token: "my-access-token"
api:
  proxy_url: ""
  tls_cert: ""
  tls_skip_verify: false
  url: https://api.safetyculture.io
db:
  connection_string: ""
  dialect: ""
export:
  inspection:
    archived: false
    completed: "true"
    included_inactive_items: false
    incremental: true
    modified_after: ""
    skip_ids: []
    template_ids: []
  path: ./export/
  tables: []
```

All of the available configuration options can be found below.

| Flag                                  | Environment Variable                            | Configuration Key                          | Description                                                                                   |
|---------------------------------------|-------------------------------------------------|--------------------------------------------|-----------------------------------------------------------------------------------------------|
| `--access-token`                      | `IAUD_ACCESS_TOKEN`                             | `access_token`                             | API Access Token                                                                              |
| `--api-url`                           | `IAUD_API_URL`                                  | `api.url`                                  | iAuditor API URL                                                                              |
| `--proxy-url`                         | `IAUD_API_PROXY_URL`                            | `api.proxy_url`                            | Proxy URL for making API requests through                                                     |
| `--tls-cert`                          | `IAUD_API_TLS_CERT`                             | `api.tls_certs`                            | Custom root CA certificate to use when making API requests                                    |
| `--tls-skip-verify`                   | `IAUD_API_TLS_SKIP_VERIFY`                      | `api.tls_skip_verify`                      | Skip verification of API TLS certificates                                                     |
| `--config-path`                       | `-`                                             | `-`                                        | config file path (default "./.iauditor-exporter.yaml")                                        |
| `--db-connection-string`              | `IAUD_DB_CONNECTION_STRING`                     | `db.connection_string`                     | Database connection string                                                                    |
| `--db-dialect`                        | `IAUD_DB_DIALECT`                               | `db.dialect`                               | Database dialect. mysql, postgres and sqlserver are the only valid options. (default "mysql") |
| `--export-path`                       | `IAUD_EXPORT_PATH`                              | `export.path`                              | CSV Export Path (default "./export/")                                                         |
| `--inspection-template-ids`           | `IAUD_EXPORT_TEMPLATE_IDS`                      | `export.template_ids`                      | Template IDs to filter inspections and schedules by (default all)                             |
| `--tables`                            | `IAUD_EXPORT_TABLES`                            | `export.tables`                            | Tables to export (default all)                                                                |
| `--inspection-archived`               | `IAUD_EXPORT_INSPECTION_ARCHIVED`               | `export.inspection.archived`               | Return archived inspections, false, true or both (default "false")                            |
| `--inspection-completed`              | `IAUD_EXPORT_INSPECTION_COMPLETED`              | `export.inspection.completed`              | Return completed inspections, false, true or both (default "both")                            |
| `--inspection-include-inactive-items` | `IAUD_EXPORT_INSPECTION_INCLUDE_INACTIVE_ITEMS` | `export.inspection.include_inactive_items` | Include inactive items in the inspection_items table (default false)                          |
| `--inspection-incremental-update`     | `IAUD_EXPORT_INSPECTION_INCREMENTAL`            | `export.inspection.incremental`            | Update inspections, inspection_items and templates tables incrementally (default true)        |
| `--inspection-skip-ids`               | `IAUD_EXPORT_INSPECTION_SKIP_IDS`               | `export.inspection.skip_ids`               | Skip storing these inspection IDs                                                             |

## Using a proxy

If you need to use a proxy to connect to the iAuditor API the following parameters can be set.

See above for the exact usage and alternative mechanisms to set them.

- `--proxy-url "http://my-proxy.corp.com"`
- `--tls-cert "/path/to/my/root-ca.pem"`
- `--tls-skip-verify true`

## Exporting

See [docs/iauditor-exporter_sql.md](docs/iauditor-exporter_sql.md) and [docs/iauditor-exporter_csv.md](docs/iauditor-exporter_csv.md) for usage of specific exporters.

## Development

To develop `iauditor-exporter` you just need the latest version of Golang which you can grab here: [https://golang.org/doc/install](https://golang.org/doc/install).

### Testing

Locally you can run `go test ./...`, this will run all of the Unit tests and Integration tests that can be run without an external DB.

SQL Database integration tests can be run by starting the SQL DBs `docker-compose up -d` and then running `make integration-tests`.

Note: these tests will be automatically when pushing or opening a pull request against the repository.
