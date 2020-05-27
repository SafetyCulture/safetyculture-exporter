# Other Databases

Support for other Databases is almost identical to SQL however there are some minor changes you should be aware of.

## MySQL

## Requirements

There is only one additional requirement, which you can install by running:

`pip install mysql`

## Config File

Here's an example config file for a MySQL connection:

```
API:
    token: API TOKEN HERE
    ssl_cert:
    proxy_http:
    proxy_https:
config_name: mysql_example
export_options:
    export_path:
    filename:
    export_archived: false
    export_completed: both
    use_real_template_name: false
    export_inactive_items: false
    export_profiles:
    template_ids:
    sync_delay_in_seconds: 70
    media_sync_offset_in_seconds: 0
    allow_table_creation: true
    merge_rows: false
    actions_merge_rows: false
    sql_table: iauditor_data
    database_type: mysql
    database_user: edd
    database_pwd: p455w0rd
    database_server: localhost
    database_port: 3306
    database_schema:
    database_name: iAuditor
```
Main things to note are the `database_type` as just `mysql` and that on `database_name` there is no need to specify a driver like with SQL.

## Postgres

## Requirements

There is only one additional requirement, which you can install by running:

`pip install psycopg2`

## Config File

Here's an example config file for a Postgres connection:

```
API:
    token: API TOKEN HERE
    ssl_cert:
    proxy_http:
    proxy_https:
config_name: postgres_example
export_options:
    export_path:
    filename:
    export_archived: false
    export_completed: both
    use_real_template_name: false
    export_inactive_items: false
    export_profiles:
    template_ids:
    sync_delay_in_seconds: 70
    media_sync_offset_in_seconds: 0
    allow_table_creation: true
    merge_rows: false
    actions_merge_rows: false
    sql_table: iauditor_data
    database_type: postgres
    database_user: edd
    database_pwd: p455w0rd
    database_server: localhost
    database_port: 5432
    database_schema:
    database_name: iAuditor
```
Main things to note are the `database_type` as just `postgres` and that on `database_name` there is no need to specify a driver like with SQL.


## Troubleshooting
MySQL and Postgres support is new and hasn't been extensively tested, however it's ran well in my tests using test databases running on Docker.

If the script has issues connecting to your server, the errors given by SQLAlchemy tend to be very useful so be sure to read the errors carefully. If you believe
there to be a bug, please raise an issue on GitHub with as much information as possible and I'll review it. 