model_config = {
    "API": {
        "token": None,
        "ssl_cert": None,
        "ssl_verify": None,
        "proxy_http": None,
        "proxy_https": None,
    },
    "config_name": "iauditor",
    "export_options": {
        "export_path": None,
        "filename": None,
        "export_archived": False,
        "export_completed": True,
        "use_real_template_name": False,
        "export_inactive_items": False,
        "preferences": None,
        "template_ids": None,
        "sync_delay_in_seconds": 900,
        "media_sync_offset_in_seconds": 900,
        "merge_rows": False,
        "actions_merge_rows": False,
        "sql_table": "iauditor_data",
        "database_type": None,
        "database_user": None,
        "database_pwd": None,
        "database_server": None,
        "database_port": None,
        "database_schema": None,
        "database_name": None,
    },
}

flat_config = [
    "token",
    "ssl_cert",
    "ssl_verify",
    "proxy_http",
    "proxy_https" "config_name",
    "export_path",
    "filename",
    "export_archived",
    "export_completed",
    "use_real_template_name",
    "export_inactive_items",
    "preferences",
    "template_ids",
    "sync_delay_in_seconds",
    "media_sync_offset_in_seconds",
    "merge_rows",
    "actions_merge_rows",
    "sql_table",
    "database_type",
    "database_user",
    "database_pwd",
    "database_server",
    "database_port",
    "database_schema",
    "database_name",
]

questions = {
    "token": {
        "type": "api_token",
        "question": "Login to generate an API token",
        "header": "",
        "parent": "API",
    },
    "proxy_http": {
        "type": "text",
        "question": "HTTP Proxy Address",
        "header": "Don" "t forget to set the HTTPS proxy, too.",
        "parent": "API",
    },
    "proxy_https": {
        "type": "text",
        "question": "HTTPS Proxy Address",
        "header": "Don" "t forget to set the HTTP proxy, too.",
        "parent": "API",
    },
    "ssl_cert": {
        "type": "text",
        "question": "Client Side Certificates",
        "header": "If you wish to apply custom SSL certificates to any requests we make, specify the path to them "
        "here. This is feature is largely untested but works by leveraging this feature of the requests "
        "module: https://requests.readthedocs.io/en/master/user/advanced/#client-side-certificates ",
        "parent": "API",
    },
    "ssl_verify": {
        "type": "text",
        "question": "SSL Cert Verification",
        "header": "If you wish to apply custom SSL certificate verification to any requests we make, specify the path "
        "to them here. This is feature is largely untested but works by leveraging this feature of the "
        "requests module: https://requests.readthedocs.io/en/master/user/advanced/#ssl-cert-verification ",
        "parent": "API",
    },
    "config_name": {
        "type": "text",
        "question": "Config Name",
        "header": """
        You can set the name of your configuration here. Very useful if you''re managing multiple 
        configurations as it''ll be used to name files and organise folders. Do not use any spaces in this name.
        """,
        "parent": None,
    },
    "export_path": {
        "type": "text",
        "question": "Export Path",
        "header": """
                    It is usually best to leave this option blank and use the default of the exports folder within the 
                    same folder where you run the tool. If you need to set the path, type it below. The path can be 
                    either absolute or relative.
                    """,
        "parent": "export_options",
    },
    "filename": {
        "type": "multi",
        "question": "When exporting PDFs, how do you want to name the files?",
        "header": """If you have complex title requirements, you should set them up on your template in iAuditor and 
        select 'Audit Title' below. Don't forget that it's possible to get two inspections with the same name if you 
        use Audit Title. Audit ID is the best option to avoid this. If you wish to set a custom ID, please edit the 
        config file directly. """,
        "options": [
            {"name": "Audit ID", "value": None},
            {"name": "Audit Title", "value": "f3245d40-ea77-11e1-aff1-0800200c9a66"},
            {"name": "Conducted By", "value": "f3245d43-ea77-11e1-aff1-0800200c9a66"},
            {"name": "Document No", "value": "f3245d46-ea77-11e1-aff1-0800200c9a66"},
            {
                "name": "Conducted At (Location)",
                "value": "f3245d44-ea77-11e1-aff1-0800200c9a66",
            },
        ],
        "parent": "export_options",
    },
    "export_archived": {
        "type": "multi",
        "question": "Archived Inspections",
        "header": """
                Select how you want to deal with archived inspections. 
                """,
        "options": [
            {
                "name": "Only export inspections that are not in the archive",
                "value": False,
            },
            {
                "name": "Export inspections both in and out of the archive",
                "value": "both",
            },
            {"name": "Only export inspections in the archive", "value": True},
        ],
        "parent": "export_options",
    },
    "export_completed": {
        "type": "multi",
        "question": "Completed Inspections",
        "header": """
                Select how you want completion status to affect your export. 
                """,
        "options": [
            {"name": "Only export inspections that are incomplete", "value": False,},
            {
                "name": "Export both complete and incomplete inspections",
                "value": "both",
            },
            {"name": "Only export completed inspections", "value": True},
        ],
        "parent": "export_options",
    },
    "use_real_template_name": {
        "type": "bool",
        "question": "Use the real template name",
        "header": """
                When exporting in CSV format, we usually export to files named after the template ID. It is
                recommended to keep things this way as it means if a template has a duplicate name, or is renamed, 
                you won't end up with clashes. Only select yes/true if you're performing a one-off export. If this
                is for on going analysis, select no/false.
                """,
        "parent": "export_options",
    },
    "export_inactive_items": {
        "type": "bool",
        "question": "Export Inactive Items",
        "header": """This setting only applies when exporting to CSV or SQL. Valid values are true (export all items) 
        or false (do not export inactive items). Items that are nested under Smart Field will be 'inactive' if the 
        smart field condition is not satisfied for these items. This option is forced to true if you're using SQL and 
        enable either of the merge_rows options. """,
        "parent": "export_options",
    },
    "preferences": {
        "type": "text",
        "question": "Set a Preference",
        "header": """
                If you wish to apply a preference to exported PDFs and Word Documents, specify the relevant 
                preference ID here. To view your preferences, re-run the tool with --list_preferences.
                """,
        "parent": "export_options",
    },
    "template_ids": {
        "type": "text",
        "question": "Export specific templates",
        "header": """
                If you do not want to export everything in your account, specify a comma separated list
                of template IDs below. 
                """,
        "parent": "export_options",
    },
    "sync_delay_in_seconds": {
        "type": "int",
        "question": "Set the sync delay in seconds",
        "header": """
                This sets the time in seconds to wait after completing one export run, before running again. 
                Defaults to 900 seconds/15 minutes. There's little benefit to going lower than a few minutes 
                each run.
                """,
        "parent": "export_options",
    },
    "media_sync_offset_in_seconds": {
        "type": "int",
        "question": "Set the media sync delay in seconds",
        "header": """
                This setting only applies to PDF and Word Exports. if a user has taken photos, it's a good idea
                to allow enough time for them to have fully synced to our servers. This delay avoids us trying
                to download a PDF/Word file immediately after it appears and potentially having missing images. 
                This defaults to 15 minutes (900 seconds), only change it if you know what you're doing. 
                """,
        "parent": "export_options",
    },
    "merge_rows": {
        "type": "bool",
        "question": "Merge SQL Rows",
        "header": """
                This setting, when set to true will update existing rows in the database when an inspection is 
                updated after being logged. There are important caveats to this option, please review the documentation
                before using this: 
                https://safetyculture.github.io/iauditor-exporter/script-setup/config/#options-only-for-sql. 
                """,
        "parent": "export_options",
    },
    "actions_merge_rows": {
        "type": "bool",
        "question": "Merge SQL Rows for Actions",
        "header": """
                This setting, when set to true will update existing rows in the database when an action is 
                updated after being logged. There are important caveats to this option, please review the documentation
                before using this: 
                https://safetyculture.github.io/iauditor-exporter/script-setup/config/#options-only-for-sql. 
                """,
        "parent": "export_options",
    },
    "sql_table": {
        "type": "text",
        "question": "Database table name",
        "header": """This is the name of the table in your database you wish to store your data in. The tool should 
        be able to create the table for you so whatever you enter here will be used as the name.""",
        "parent": "export_options",
    },
    "database_type": {
        "type": "multi",
        "question": "Database Type",
        "header": """Set the type of database you're using.""",
        "options": [
            {"name": "SQL", "value": "mssql"},
            {"name": "MySQL", "value": "mysql"},
            {"name": "Postgres", "value": "postgres"},
        ],
        "parent": "export_options",
    },
    "database_user": {
        "type": "text",
        "question": "Database user",
        "header": """Set the username to use when accessing the database.""",
        "parent": "export_options",
    },
    "database_pwd": {
        "type": "password",
        "question": "Database password",
        "header": """Set the password to use when accessing the database.""",
        "parent": "export_options",
    },
    "database_server": {
        "type": "text",
        "question": "Database server",
        "header": """
        This option sets the address of your database server. e.g. localhost or 192.168.1.88, for example. 
        Remember to ensure your database allows remote connections if you're not running the tool on the same box as your database.
        """,
        "parent": "export_options",
    },
    "database_port": {
        "type": "int",
        "question": "Database port",
        "header": """
    Specify the port to access your database. For reference, default ports are: 
    SQL: 1433
    MySQL: 3308
    Postgres: 5432
    """,
        "parent": "export_options",
    },
    "database_schema": {
        "type": "text",
        "question": "Database schema",
        "header": """
Specify the schema you wish to use. The defaults are: SQL: dbo, MySQL: From tests, this can usually be left blank, 
Postgres: public.
""",
        "parent": "export_options",
    },
    "database_name": {
        "type": "text",
        "question": "Database name",
        "header": """
        This option specifies the name of the database you are using. For Postgres and MySQL, the name 
        alone is enough. 
        
        For SQL, you must specify the driver you wish to use. To do this, type your database name 
        and follow it with ?driver=ODBC Driver 17 for SQL Server (or other driver depending on your requirements. 
        For example, if your database name was 'iAuditor', you'd enter: iAuditor?driver=ODBC Driver 17 for SQL Server 
        """,
        "parent": "export_options",
    },
}
