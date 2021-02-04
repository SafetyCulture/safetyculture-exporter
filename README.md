# iAuditor Exporter

[![Maintainability](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/maintainability)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/test_coverage)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/test_coverage)

The iAuditor Exporter is a CLI tool for extracting your iAuditor data that’s available to all our Premium customers. It’s used to bulk export `inspections`, `templates`, `sites`, `users`, `groups` and `schedules` to CSV and SQL formats for business intelligence tools. It’s also capable of exporting inspections to formats such as JSON, PDF and Word.

- [Before you begin](#before-you-begin)
- [Install](#install)
- [Configure](#configure)
  - [Using a proxy](#using-a-proxy)
- [Usage](#usage)
  - [Exporting to CSV](#exporting-to-csv)
  - [Exporting to SQL](#exporting-to-sql)
    - [MySQL](#mysql)
    - [PostgreSQL](#postgresql)
    - [SQL Server](#sql-server)
  - [Exporting inspections to JSON](#exporting-inspections-to-json)
  - [Listing Schemas](#listing-schemas)
- [Old version (Python)](#old-version-python)
  - [Migrating to new version](#migrating-to-new-version)
- [Development](#development)

## Before you begin

Please note that you must be on our [Premium subscription](https://safetyculture.com/pricing/) in order to install and run the iAuditor Exporter tool.

We recommend that you have some basic knowledge of running command lines before installing the iAuditor Exporter CLI tool.

## Install

Download the latest release from [iauditor-exporter/releases](https://github.com/SafetyCulture/iauditor-exporter/releases) and extract it to a directory on you machine.

Open the terminal and navigate to the folder the archive was extracted to. Run `./iauditor-exporter --help` to make sure your installation was successful.

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

All of the available configuration options and flags can be found below.

#### `access_token`
API Access Token
> flag: `--access-token`\
env: `IAUD_ACCESS_TOKEN`

#### `api.url`
iAuditor API URL
> flag: `--api-url`\
env: `IAUD_API_URL`\
default: `https://api.safetyculture.io`

#### `api.proxy_url`
Proxy URL for making API requests through
> flag: `--proxy-url`\
env: `IAUD_API_PROXY_URL`

#### `api.tls_certs`
Custom root CA certificate to use when making API requests
> flag: `--tls-cert`\
env: `IAUD_API_TLS_CERT`

#### `api.tls_skip_verify`
Skip verification of API TLS certificates
> flag: `--tls-skip-verify`\
env: `IAUD_API_TLS_SKIP_VERIFY`\
default: `false`

#### `db.connection_string`
Database connection string
> flag: `--db-connection-string`\
env: `IAUD_DB_CONNECTION_STRING`

#### `db.dialect`
Database dialect. mysql, postgres and sqlserver are the only valid options.
> flag: `--db-dialect`\
env: `IAUD_DB_DIALECT`\
default: `mysql`

#### `export.path`
Export path on disk
> flag: `--export-path`\
env: `IAUD_EXPORT_PATH`\
default: `./export/`

#### `export.template_ids`
Template IDs to filter inspections and schedules
> flag: `--template-ids`\
env: `IAUD_EXPORT_TEMPLATE_IDS`\
default: All templates

#### `export.tables`
Tables to export. inspections, inspection_items, templates, template_permissions, sites, users, groups, group_users, schedules, schedule_assignees and schedule_occurrences are only valid values.
> flag: `--tables`\
env: `IAUD_EXPORT_TABLES`\
default: All tables

#### `export.incremental`
Update actions, action_assignees, inspections, inspection_items and templates tables incrementally
> flag: `--incremental-update`\
env: `IAUD_EXPORT_INCREMENTAL`\
default: `true`

#### `export.modified_after`
Return inspections and actions modified only after this date
> flag: `--modified-after`\
env: `IAUD_EXPORT_MODIFIED_AFTER`
formats:

- 2006-01-02T15:04:05Z07:00 (ISO8601)
- Mon, 02 Jan 2006 15:04:05 MST (RFC1123)
- 02 Jan 06 15:04 MST (RFC822)
- Monday, 02-Jan-06 15:04:05 MST (RFC850)
- Mon Jan 2 15:04:05 2006 (ANSIC)
- Mon Jan 2 15:04:05 MST 2006
- 2006-01-02
- 02 Jan 2006

#### `export.inspection.archived`
Return archived inspections, false, true or both
> flag: `--inspection-archived`\
env: `IAUD_EXPORT_INSPECTION_ARCHIVED`\
default: `false`

#### `export.inspection.completed`
Return completed inspections, false, true or both
> flag: `--inspection-completed`\
env: `IAUD_EXPORT_INSPECTION_COMPLETED`\
default: `both`

#### `export.inspection.include_inactive_items`
Include inactive items in the inspection_items table
> flag: `--inspection-include-inactive-items`\
env: `IAUD_EXPORT_INSPECTION_INCLUDE_INACTIVE_ITEMS`\
default: `false`

#### `export.inspection.skip_ids`
Skip storing these inspection IDs
> flag: `--inspection-skip-ids`\
env: `IAUD_EXPORT_INSPECTION_SKIP_IDS`

### Using a proxy

If you need to use a proxy to connect to the iAuditor API the following parameters can be set.

See above for the exact usage and alternative mechanisms to set them.

- `--proxy-url "http://my-proxy.corp.com"`
- `--tls-cert "/path/to/my/root-ca.pem"`
- `--tls-skip-verify true`

## Usage

See [docs/iauditor-exporter.md](docs/iauditor-exporter.md) for list of commands and options.

### Exporting to CSV

See [docs/iauditor-exporter_csv.md](docs/iauditor-exporter_csv.md) for usage and options.

### Exporting to SQL DBs

iAuditor Exporter supports exporting data to MySQL, PostgreSQL and SQL Server databases. See [docs/iauditor-exporter_sql.md](docs/iauditor-exporter_sql.md) for usage and options.

#### MySQL

```
iauditor-exporter sql --db-dialect mysql --db-connection-string user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
```

Please refer to [this page](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for supported formats for the connection string.

#### PostgreSQL

```
iauditor-exporter sql --db-dialect postgres --db-connection-string postgresql://user:pass@localhost:5434/dbname
```

#### SQL Server

```
iauditor-exporter sql --db-dialect sqlserver --db-connection-string sqlserver://user:pass@localhost:1433?database=dbname
```

Please refer to [this page](https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn) for supported formats for the connection string.


### Inspection report export

iAuditor exporter supports generating PDF and Word reports for inspections. See [docs/iauditor-exporter_report.md](docs/iauditor-exporter_report.md) for usage and options.

### Exporting inspections to JSON

See [docs/iauditor-exporter_inspection-json.md](docs/iauditor-exporter_inspection-json.md) for usage and options.

### Listing Schemas

You can list all available tables with their schemas using following command.

```
iauditor-exporter schema
```

#### Relationships

![Relationships Diagram](docs/assets/relationships.png)

## Old version (Python)

Old version of this tool needs Python to be installed on the target machine.
We no longer support this version but it can be accessed from [this location](https://github.com/SafetyCulture/iauditor-exporter/tree/v2).

### Migrating to new version

We highly recommend to upgrade to the new version of iAuditor Exporter tool as soon as possible.
There is feature parity between the new version and old version of this tool.

Although configuration values are similar, the format of the configuration files varies between two versions.
Please refer to the [Configure](#configure) section of this page to learn more about the format of the new configuration file.

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
