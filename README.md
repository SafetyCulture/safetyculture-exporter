# iAuditor Exporter

[![Maintainability](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/maintainability)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/39eecd9ef3573ecca044/test_coverage)](https://codeclimate.com/github/SafetyCulture/iauditor-exporter/test_coverage)

iAuditor Exporter is a command-line tool (CLI tool) that’s available to all our Premium and Enterprise customers. You can use the iAuditor Exporter to export your iAuditor data, such as inspections, templates, schedules, and actions to multiple formats that can be used for business intelligence tools or record keeping.

For instructions on downloading and running the iAuditor Exporter, as well as interpreting the data output, please check out our [iAuditor Exporter wiki](https://github.com/SafetyCulture/iauditor-exporter/wiki).

* [iAuditor Exporter wiki](https://github.com/SafetyCulture/iauditor-exporter/wiki/Home)
* [Run the iAuditor Exporter](https://github.com/SafetyCulture/iauditor-exporter/wiki/Run-the-iAuditor-Exporter)
  * [Download](https://github.com/SafetyCulture/iauditor-exporter/wiki/Download-the-iAuditor-Exporter)
  * [Configuration](https://github.com/SafetyCulture/iauditor-exporter/wiki/Configure-the-iAuditor-Exporter)
  * [Database support](https://github.com/SafetyCulture/iauditor-exporter/wiki/Database-support)
  * [Export data](https://github.com/SafetyCulture/iauditor-exporter/wiki/Export-data)
  * [Errors](https://github.com/SafetyCulture/iauditor-exporter/wiki/Errors)
* [Understand the data](https://github.com/SafetyCulture/iauditor-exporter/wiki/Understand-the-data)
  * [CSV or SQL?](https://github.com/SafetyCulture/iauditor-exporter/wiki/CSV-or-SQL%3F)
  * [Data columns](https://github.com/SafetyCulture/iauditor-exporter/wiki/Data-columns)
    * [Inspections](https://github.com/SafetyCulture/iauditor-exporter/wiki/Inspections-data-sets)
    * [Templates](https://github.com/SafetyCulture/iauditor-exporter/wiki/Templates-data-sets)
    * [Organization](https://github.com/SafetyCulture/iauditor-exporter/wiki/Organization-data-sets)
    * [Schedules](https://github.com/SafetyCulture/iauditor-exporter/wiki/Schedules-data-sets)
    * [Actions](https://github.com/SafetyCulture/iauditor-exporter/wiki/Actions-data-sets)
    * [Issues](https://github.com/SafetyCulture/iauditor-exporter/wiki/Issues-data-sets)

> The [Python version of the iAuditor Exporter](https://github.com/SafetyCulture/iauditor-exporter/tree/v2) is no longer being maintained. We recommend downloading this latest version for faster exporting and additional data sets.

***

## Development

To develop the `iauditor-exporter`, you'll need the [latest version of Golang](https://golang.org/doc/install).

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
