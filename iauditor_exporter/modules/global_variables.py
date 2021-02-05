import logging
import os
import sys

sys.path.append(os.path.join(os.path.dirname(__file__), "..", ".."))

# Possible values here are DEBUG, INFO, WARN, ERROR and CRITICAL
LOG_LEVEL = logging.DEBUG

# Stores the API access token and other configuration settings
DEFAULT_CONFIG_FILENAME = "config.yaml"

# Wait 15 minutes by default between sync attempts
DEFAULT_SYNC_DELAY_IN_SECONDS = 900

# Only download audits older than 10 minutes
DEFAULT_MEDIA_SYNC_OFFSET_IN_SECONDS = 600

# The file that stores the "date modified" of the last successfully synced audit
SYNC_MARKER_FILENAME = "last_successful/last_successful.txt"

# The file that stores the ISO date/time string of the last successful actions export
ACTIONS_SYNC_MARKER_FILENAME = "last_successful/last_successful_actions_export.txt"

# the file that stores all exported actions in CSV format
ACTIONS_EXPORT_FILENAME = "iauditor_actions.csv"

# Whether to export inactive items to CSV
DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV = True

# When exporting actions to CSV, if property is None, print this value to CSV
EMPTY_RESPONSE = ""

# Not all Audits will actually contain an Audit Title item. For examples, when Audit Title rules are set, the Audit
# Title item is not going to be included by default.
# When this item ID is specified in the custom export filename configuration, the audit_data.name property will
# be used to populate the data as it covers all cases.
AUDIT_TITLE_ITEM_ID = "f3245d40-ea77-11e1-aff1-0800200c9a66"

# Properties kept in settings dictionary which takes its values from config.YAML
API_TOKEN = "api_token"
HEROKU_URL = "heroku_url"
SSL_CERT = "ssl_cert"
SSL_VERIFY = "ssl_verify"
PROXY_HTTP = "proxy_http"
PROXY_HTTPS = "proxy_https"
CONFIG_NAME = "config_name"
EXPORT_PATH = "export_path"
PREFERENCES = "preferences"
FILENAME_ITEM_ID = "filename_item_id"
SYNC_DELAY_IN_SECONDS = "sync_delay_in_seconds"
EXPORT_INACTIVE_ITEMS_TO_CSV = "export_inactive_items_to_csv"
MEDIA_SYNC_OFFSET_IN_SECONDS = "media_sync_offset_in_seconds"
EXPORT_FORMATS = "export_formats"
TEMPLATE_IDS = "template_ids"
SQL_TABLE = "sql_table"
DB_TYPE = "database_type"
DB_USER = "database_user"
DB_PWD = "database_pwd"
DB_SERVER = "database_server"
DB_PORT = "database_port"
DB_NAME = "database_name"
DB_SCHEMA = "database_schema"
USE_REAL_TEMPLATE_NAME = "use_real_template_name"
EXPORT_ARCHIVED = "export_archived"
EXPORT_COMPLETED = "export_completed"
MERGE_ROWS = "merge_rows"
ALLOW_TABLE_CREATION = "allow_table_creation"
ACTIONS_TABLE = "actions_table"
ACTIONS_MERGE_ROWS = "actions_merge_rows"

# Used to create a default config file for new users
DEFAULT_CONFIG_FILE_YAML = [
    "API:",
    "\n    token: ",
    "\nconfig_name: " "\nexport_options: ",
    "\n    export_path: ",
    "\n    export_archived: false",
    "\n    export_completed: true",
    "\n    use_real_template_name: false" "\n    filename: ",
    "\n    export_inactive_items: false",
    "\n    preferences: ",
    "\n    sync_delay_in_seconds: 300",
    "\n    media_sync_offset_in_seconds: ",
    "\n    template_ids: ",
    "\n    merge_rows: false",
    "\n    actions_merge_rows: false",
    "\n    allow_table_creation: false",
    "\n    sql_table: ",
    "\n    database_type: ",
    "\n    database_server: ",
    "\n    database_user: ",
    "\n    database_pwd: ",
    "\n    database_port: ",
    "\n    database_name: DB-NAME?driver=ODBC Driver 17 for SQL Server",
    "\n    database_schema: ",
]
