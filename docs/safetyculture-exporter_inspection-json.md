## safetyculture-exporter inspection-json

Export SafetyCulture inspections to json files

```
safetyculture-exporter inspection-json [flags]
```

### Examples

```
// Limit inspections to these templates
safetyculture-exporter inspection-json --template-ids template_F492E54D87F2419E9398F7BDCA0FA5D9,template_d54e06808d2f11e2893e83a731dba0ca

// Customise export location
safetyculture-exporter inspection-json --export-path /path/to/export/to
```

### Options

```
  -t, --access-token string                 API Access Token
      --action-limit int                    Number of actions fetched at once. Lower this number if the exporter fails to load the data (default 100)
      --api-url string                      API URL (default "https://api.safetyculture.io")
      --export-path string                  File Export Path (default "./export/")
  -h, --help                                help for inspection-json
      --incremental                         Update inspections, inspection_items and templates tables incrementally (default true)
      --inspection-archived string          Return archived inspections, false, true or both (default "false")
      --inspection-completed string         Return completed inspections, false, true or both (default "true")
      --inspection-include-inactive-items   Include inactive items in the inspection_items table (default false)
      --inspection-limit int                Number of inspections fetched at once. Lower this number if the exporter fails to load the data (default 100)
      --inspection-skip-ids strings         Skip storing these inspection IDs
      --inspection-web-report-link string   Web report link format. Can be public or private (default "private")
      --modified-after string               Return inspections modified after this date (see readme for supported formats)
      --proxy-url string                    Proxy URL for making API requests through
      --template-ids strings                Template IDs to filter inspections and schedules by (default all)
      --tls-cert string                     Custom root CA certificate to use when making API requests
      --tls-skip-verify                     Skip verification of API TLS certificates
```

### Options inherited from parent commands

```
      --config-path string   config file (default "./safetyculture-exporter.yaml")
```

### SEE ALSO

* [safetyculture-exporter](safetyculture-exporter.md)	 - A CLI tool for extracting your SafetyCulture data

###### Auto generated by spf13/cobra on 21-Oct-2021
