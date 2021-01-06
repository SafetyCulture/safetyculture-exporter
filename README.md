# iAuditor Exporter

[![Maintainability](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/maintainability)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/test_coverage)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/test_coverage)

iAuditor Exporter is a CLI tool for extracting your iAuditor data.

To see the full set of commands available run `iauditor-exporter --help`

## Install

Download the latest release from [iauditor-exporter/releases](https://github.com/SafetyCulture/iauditor-exporter/releases).

### Quick Start

To get up and running quickly run the `configure` command to create a stub config file.

`./iauditor-exporter configure`

This will create a config file `iauditor-exporter.yaml` in the same directory. Use `--config-path "/my/path/iauditor-exporter.yaml"` to set a custom path.

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
  dialect: mysql
export:
  inspection:
    archived: "false"
    completed: both
    included_inactive_items: false
    incremental: true
    skip_ids: []
  path: ./export/
  tables: []
  template_ids: []
```

All of the available configuration options can be found below.

### `access_token`
API Access Token  
> Flag: `--access-token`  
Env: `IAUD_ACCESS_TOKEN`  


### `api.url`
iAuditor API URL  
> Flag: `--api-url`  
Env: `IAUD_API_URL`  
Default: `https://api.safetyculture.io`

| Configuration Key                          | Description                                                                                   |
|--------------------------------------------|-----------------------------------------------------------------------------------------------|
| `access_token`                             | API Access Token                                                                              |
| `api.url`                                  | iAuditor API URL                                                                              |
| `api.proxy_url`                            | Proxy URL for making API requests through                                                     |
| `api.tls_certs`                            | Custom root CA certificate to use when making API requests                                    |
| `api.tls_skip_verify`                      | Skip verification of API TLS certificates                                                     |
| `db.connection_string`                     | Database connection string                                                                    |
| `db.dialect`                               | Database dialect. mysql, postgres and sqlserver are the only valid options. (default "mysql") |
| `export.path`                              | CSV Export Path (default "./export/")                                                         |
| `export.template_ids`                      | Template IDs to filter inspections and schedules by (default all)                             |
| `export.tables`                            | Tables to export (default all)                                                                |
| `export.inspection.archived`               | Return archived inspections, false, true or both (default "false")                            |
| `export.inspection.completed`              | Return completed inspections, false, true or both (default "both")                            |
| `export.inspection.include_inactive_items` | Include inactive items in the inspection_items table (default false)                          |
| `export.inspection.incremental`            | Update inspections, inspection_items and templates tables incrementally (default true)        |
| `export.inspection.skip_ids`               | Skip storing these inspection IDs                                                             |

## Using a proxy

If you need to use a proxy to connect to the iAuditor API the following parameters can be set.

See above for the exact usage and alternative mechanisms to set them.

- `--proxy-url "http://my-proxy.corp.com"`
- `--tls-cert "/path/to/my/root-ca.pem"`
- `--tls-skip-verify true`

## Exporting to CSV

See [docs/iauditor-exporter_csv.md](docs/iauditor-exporter_csv.md) for usage and options.

## Exporting to SQL DBs

iAuditor Exporter supports exporting data to MySQL, PostgreSQL and SQL Server databases. See [docs/iauditor-exporter_sql.md](docs/iauditor-exporter_sql.md) for usage and options.

### MySQL

```
iauditor-exporter sql --db-dialect mysql --db-connection-string user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

Please refer to [this page](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for supported formats for the connection string.

### PostgreSQL

```
iauditor-exporter sql --db-dialect postgres --db-connection-string postgresql://user:pass@localhost:5434/dbname
```

### SQL Server

```
iauditor-exporter sql --db-dialect sqlserver --db-connection-string sqlserver://user:pass@localhost:1433?database=dbname
```

Please refer to [this page](https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn) for supported formats for the connection string.

## Listing Schemas

You can list all available tables with their schemas using following command.

```
iauditor-exporter schema
```

## Development

To develop `iauditor-exporter` you just need the latest version of Golang which you can grab here: [https://golang.org/doc/install](https://golang.org/doc/install).

### Testing

Locally you can run `go test ./...`, this will run all of the Unit tests and Integration tests that can be run without an external DB.

SQL Database integration tests can be run by starting the SQL DBs `docker-compose up -d` and then running `make integration-tests`.

Note: these tests will be automatically when pushing or opening a pull request against the repository.

### Releasing

To release a new version you need just need to push a new tag to GitHub and [goreleaser](https://goreleaser.com) will do most of the work.

1. Checkout the `main` branch and pull the latest changes. If you don't you'll tag the wrong commit for release!
2. Create your tag, make sure it follows [Semantic Versioning](https://semver.org) and increments on the [latest release](https://github.com/SafetyCulture/iauditor-exporter/releases)\
`git tag -a v3.0.0 -m "Initial Public Release"`.\
Acceptable versions include `v3.0.0`, `v3.0.0-alpha.22`, `v3.0.0-prealpha.22`, `v3.0.0-beta.22`.
3. Push your tag to GitHub\
`git push origin v3.0.0`
4. Update the [release draft](https://github.com/SafetyCulture/iauditor-exporter/releases) and publish it!
