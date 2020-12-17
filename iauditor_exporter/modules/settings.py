import argparse
import re

import questionary as questionary
import yaml
from questionary import Separator
from rich import print, box
from rich.panel import Panel
from safetypy import safetypy as sp
from yaml.scanner import ScannerError

from iauditor_exporter.modules.global_variables import *
from iauditor_exporter.modules.last_successful import (
    get_last_successful,
    parse_ls,
    update_sync_marker_file,
)
from iauditor_exporter.modules.logger import (
    log_critical_error,
    create_directory_if_not_exists,
)
from iauditor_exporter.modules.setup_questions import questions, model_config
from iauditor_exporter.modules.sql import test_sql_settings


def load_setting_api_access_token(logger, config_settings):
    """
    Attempt to parse API token from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 API token if valid, else None
    """
    try:
        api_token = config_settings["API"]["token"]
        token_is_valid = re.match("^[a-f0-9]{64}$", api_token)
        if token_is_valid:
            logger.debug("API token matched expected pattern")
            return api_token
        else:
            logger.error("API token failed to match expected pattern")
            return None
    except Exception as ex:
        log_critical_error(logger, ex, "Exception parsing API token from config.yaml")
        return None


def docker_load_setting_api_access_token(logger, api_token):
    """
    Attempt to parse API token from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 API token if valid, else None

    Args:
        api_token:
    """
    try:
        token_is_valid = re.match("^[a-f0-9]{64}$", api_token)
        if token_is_valid:
            logger.debug("API token matched expected pattern")
            return api_token
        else:
            logger.error("API token failed to match expected pattern")
            return None
    except Exception as ex:
        log_critical_error(logger, ex, "Exception parsing API token from config.yaml")
        return None


def load_export_inactive_items_to_csv(logger, config_settings):
    """
    Attempt to parse export_inactive_items from config settings. Value of true or false is expected.
    True means the CSV exporter will include inactive items. False means the CSV exporter will exclude inactive items.
    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 value of export_inactive_items_to_csv if valid, else DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV
    """
    try:
        if config_settings["export_options"]["merge_rows"] is True:
            logger.info(
                "Merge rows is enabled, turning on the export of inactive items."
            )
            export_inactive_items_to_csv = True
        else:
            export_inactive_items_to_csv = config_settings["export_options"][
                "export_inactive_items"
            ]
            if not isinstance(export_inactive_items_to_csv, bool):
                logger.info(
                    "Invalid export_inactive_items value from configuration file, defaulting to true"
                )
                export_inactive_items_to_csv = DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV
        return export_inactive_items_to_csv
    except Exception as ex:
        log_critical_error(
            logger,
            ex,
            "Exception parsing export_inactive_items from the configuration file, defaulting to {0}".format(
                str(DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV)
            ),
        )
        return DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV


def load_setting_sync_delay(logger, config_settings):
    """
    Attempt to parse delay between sync loops from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 extracted sync delay if valid, else DEFAULT_SYNC_DELAY_IN_SECONDS
    """
    try:
        sync_delay = config_settings["export_options"]["sync_delay_in_seconds"]
        sync_delay_is_valid = re.match("^[0-9]+$", str(sync_delay))
        if sync_delay_is_valid and sync_delay >= 0:
            if sync_delay < DEFAULT_SYNC_DELAY_IN_SECONDS:
                "{0} seconds".format(
                    logger.info(
                        "Sync delay is less than the minimum recommended value of "
                        + str(DEFAULT_SYNC_DELAY_IN_SECONDS)
                    )
                )
            return sync_delay
        else:
            logger.info(
                "Invalid sync_delay_in_seconds from the configuration file, defaulting to {0}".format(
                    str(DEFAULT_SYNC_DELAY_IN_SECONDS)
                )
            )
            return DEFAULT_SYNC_DELAY_IN_SECONDS
    except Exception as ex:
        log_critical_error(
            logger,
            ex,
            "Exception parsing sync_delay from the configuration file, defaulting to {0}".format(
                str(DEFAULT_SYNC_DELAY_IN_SECONDS)
            ),
        )
        return DEFAULT_SYNC_DELAY_IN_SECONDS


def load_setting_preference_mapping(logger, config_settings):
    """
    Attempt to parse preference settings from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 export preference mapping if valid, else None
    """
    try:
        preference_mapping = {}
        preference_settings = config_settings["export_options"]["preferences"]
        if preference_settings is not None:
            preference_lines = preference_settings.split(" ")
            for preference in preference_lines:
                template_id = preference[: preference.index(":")]
                if template_id not in preference_mapping.keys():
                    preference_mapping[template_id] = preference
        return preference_mapping
    except KeyError:
        logger.debug("No preference key in the configuration file")
        return None
    except Exception as ex:
        log_critical_error(
            logger, ex, "Exception getting preferences from the configuration file"
        )
        return None


def load_setting_export_path(logger, config_settings):
    """
    Attempt to extract export path from config settings

    :param config_settings:  config settings loaded from config file
    :param logger:           the logger
    :return:                 export path, None if path is invalid or missing
    """
    try:
        export_path = config_settings["export_options"]["export_path"]
        if export_path is not None:
            return export_path
        else:
            return None
    except Exception as ex:
        log_critical_error(
            logger, ex, "Exception getting export path from the configuration file"
        )
        return None


def load_setting_media_sync_offset(logger, config_settings):
    """

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 media sync offset parsed from file, else default media sync offset
                             defined as global constant
    """
    try:
        media_sync_offset = config_settings["export_options"][
            "media_sync_offset_in_seconds"
        ]
        if (
            media_sync_offset is None
            or media_sync_offset < 0
            or not isinstance(media_sync_offset, int)
        ):
            media_sync_offset = DEFAULT_MEDIA_SYNC_OFFSET_IN_SECONDS
        return media_sync_offset
    except Exception as ex:
        log_critical_error(
            logger, ex, "Exception parsing media sync offset from config file"
        )
        return DEFAULT_MEDIA_SYNC_OFFSET_IN_SECONDS


def parse_export_filename(audit_json, filename_item_id):
    """
    Get 'response' value of specified header item to use for export file name

    :param header_items:      header_items array from audit JSON
    :param filename_item_id:  item_id from config settings
    :return:                  'response' value of specified item from audit JSON
    """
    if filename_item_id is None:
        return None
    # Not all Audits will actually contain an Audit Title item. For examples, when Audit Title rules are set,
    # the Audit Title item is not going to be included by default.
    # When this item ID is specified in the custom export filename configuration, the audit_data.name property
    # will be used to populate the data as it covers all cases.
    if (
        filename_item_id == AUDIT_TITLE_ITEM_ID
        and "audit_data" in audit_json.keys()
        and "name" in audit_json["audit_data"].keys()
    ):
        return audit_json["audit_data"]["name"].replace("/", "_")
    for item in audit_json["header_items"]:
        if item["item_id"] == filename_item_id:
            if "responses" in item.keys():
                if (
                    "text" in item["responses"].keys()
                    and item["responses"]["text"].strip() != ""
                ):
                    return item["responses"]["text"]
    return None


def get_filename_item_id(logger, config_settings):
    """
    Attempt to parse item_id for file naming from config settings

    :param logger:          the logger
    :param config_settings: config settings loaded from config file
    :return:                item_id extracted from config_settings if valid, else None
    """

    try:
        filename_item_id = config_settings["export_options"]["filename"]
        if filename_item_id is not None:
            if len(filename_item_id) > 36:
                logger.critical(
                    "You can only specify one value for the filename. Please remove any additional item "
                    "IDs and try again. For more complex title rules, consider setting the title rules "
                    "within iAuditor. Defaulting to Audit ID."
                )
            if filename_item_id == "f3245d42-ea77-11e1-aff1-0800200c9a66":
                logger.critical(
                    "Date fields are not compatible with the title rule feature. Defaulting to Audit ID"
                )
            else:
                return filename_item_id
        else:
            return None
    except Exception as ex:
        log_critical_error(
            logger,
            ex,
            'Exception retrieving setting "filename" from the configuration file',
        )
        return None


def set_env_defaults(name, env_var, logger):
    # if env_var is None or '':
    if name == "DB_SERVER" and env_var == "heroku":
        return os.environ["DATABASE_URL"]
    if not env_var:
        if name == "CONFIG_NAME":
            logger.error("You must set the CONFIG_NAME")
            sys.exit()
        elif name == "DB_SCHEMA":
            env_var = "dbo"
        elif name.startswith("DB_"):
            env_var = None
        elif name == "SQL_TABLE":
            env_var = None
        elif name == "TEMPLATE_IDS":
            env_var = None
        else:
            env_var = "false"
    if env_var == "None":
        env_var = None
    print(name, " set to ", env_var)
    return env_var


def load_setting_ssl_cert(logger, config_settings):
    cert_location = None
    if "ssl_cert" in config_settings["API"]:
        if config_settings["API"]["ssl_cert"]:
            cert_location = config_settings["API"]["ssl_cert"]
    return cert_location


def load_setting_ssl_verify(logger, config_settings):
    verify_cert = None
    if "ssl_verify" in config_settings["API"]:
        if config_settings["API"]["ssl_verify"]:
            verify_cert = config_settings["API"]["ssl_verify"]
    return verify_cert


def load_setting_proxy(logger, config_settings, http_or_https):
    proxy = None
    if http_or_https == "https":
        if "proxy_https" in config_settings["API"]:
            if config_settings["API"]["proxy_https"]:
                proxy = config_settings["API"]["proxy_https"]
    elif http_or_https == "http":
        if "proxy_http" in config_settings["API"]:
            if config_settings["API"]["proxy_http"]:
                proxy = config_settings["API"]["proxy_http"]
    else:
        proxy = None
    return proxy


def load_actions_table(actions_table_name):
    if actions_table_name is None:
        actions_table_name = "iauditor"
        return actions_table_name
    else:
        return actions_table_name


def load_config_settings(logger, path_to_config_file, docker_enabled=False):
    """
    Load config settings from config file

    :param logger:              the logger
    :param path_to_config_file: location of config file
    :return:                    settings dictionary containing values for:
                                api_token, export_path, preferences,
                                filename_item_id, sync_delay_in_seconds loaded from
                                config file, media_sync_offset_in_seconds
    """

    if docker_enabled is True:
        settings = {
            API_TOKEN: docker_load_setting_api_access_token(
                logger, os.environ["API_TOKEN"]
            ),
            HEROKU_URL: os.environ.get("HEROKU_POSTGRESQL_OLIVE_URL"),
            SSL_CERT: set_env_defaults("SSL_CERT", os.environ["SSL_CERT"], logger),
            SSL_VERIFY: set_env_defaults(
                "SSL_VERIFY", os.environ["SSL_VERIFY"], logger
            ),
            PROXY_HTTP: set_env_defaults(
                "PROXY_HTTP", os.environ["PROXY_HTTP"], logger
            ),
            PROXY_HTTPS: set_env_defaults(
                "PROXY_HTTPS", os.environ["PROXY_HTTPS"], logger
            ),
            EXPORT_PATH: None,
            # PREFERENCES: load_setting_preference_mapping(logger, config_settings),
            # FILENAME_ITEM_ID: get_filename_item_id(logger, config_settings),
            SYNC_DELAY_IN_SECONDS: int(os.environ["SYNC_DELAY_IN_SECONDS"]),
            # EXPORT_INACTIVE_ITEMS_TO_CSV: load_export_inactive_items_to_csv(logger, config_settings),
            MEDIA_SYNC_OFFSET_IN_SECONDS: int(
                os.environ["MEDIA_SYNC_OFFSET_IN_SECONDS"]
            ),
            TEMPLATE_IDS: set_env_defaults(
                "TEMPLATE_IDS", os.environ["TEMPLATE_IDS"], logger
            ),
            SQL_TABLE: set_env_defaults("SQL_TABLE", os.environ["SQL_TABLE"], logger),
            DB_TYPE: set_env_defaults("DB_TYPE", os.environ["DB_TYPE"], logger),
            DB_USER: set_env_defaults("DB_USER", os.environ["DB_USER"], logger),
            DB_PWD: set_env_defaults("DB_PWD", os.environ["DB_PWD"], logger),
            DB_SERVER: set_env_defaults("DB_SERVER", os.environ["DB_SERVER"], logger),
            DB_PORT: set_env_defaults("DB_PORT", os.environ["DB_PORT"], logger),
            DB_NAME: set_env_defaults("DB_NAME", os.environ["DB_NAME"], logger),
            DB_SCHEMA: set_env_defaults("DB_SCHEMA", os.environ["DB_SCHEMA"], logger),
            USE_REAL_TEMPLATE_NAME: set_env_defaults(
                "USE_REAL_TEMPLATE_NAME", os.environ["USE_REAL_TEMPLATE_NAME"], logger
            ),
            CONFIG_NAME: set_env_defaults(
                "CONFIG_NAME", os.environ["CONFIG_NAME"], logger
            ),
            EXPORT_ARCHIVED: set_env_defaults(
                "EXPORT_ARCHIVED", os.environ["EXPORT_ARCHIVED"], logger
            ),
            EXPORT_COMPLETED: set_env_defaults(
                "EXPORT_COMPLETED", os.environ["EXPORT_COMPLETED"], logger
            ),
            MERGE_ROWS: set_env_defaults(
                "MERGE_ROWS", os.environ["MERGE_ROWS"], logger
            ),
            ALLOW_TABLE_CREATION: set_env_defaults(
                "ALLOW_TABLE_CREATION", os.environ["ALLOW_TABLE_CREATION"], logger
            ),
            ACTIONS_TABLE: "iauditor_actions_data",
            ACTIONS_MERGE_ROWS: set_env_defaults(
                "ACTIONS_MERGE_ROWS", os.environ["ACTIONS_MERGE_ROWS"], logger
            ),
            PREFERENCES: None,
            FILENAME_ITEM_ID: None,
            EXPORT_INACTIVE_ITEMS_TO_CSV: None,
        }
    else:
        try:
            config_settings = yaml.safe_load(open(path_to_config_file))
        except ScannerError as e:
            logger.error(e)
            logger.critical(
                "There is a problem with your config file. The most likely reason is not leaving spaces "
                "after the colons. Open your config.yaml file and ensure that after every : you have left "
                "a space. For example, config_name:iauditor would create this error, it should be "
                "config_name: iauditor "
            )
            logger.critical(
                "Please refer to "
                "https://safetyculture.github.io/iauditor-exporter/script-setup/config/ for more "
                "information."
            )
            sys.exit()
        if config_settings["config_name"] is None:
            logger.info("The Config Name has been left blank, defaulting to iauditor.")
            config_name = "iauditor"
        elif " " in config_settings["config_name"]:
            config_name = config_settings["config_name"].replace(" ", "_")
        else:
            config_name = config_settings["config_name"]

        if re.match("^[A-Za-z0-9_-]*$", config_name):
            config_name = config_name
        else:
            logger.critical(
                "Config name can only contain letters, numbers, hyphens or underscores."
            )
            sys.exit()
        if "allow_table_creation" in config_settings["export_options"]:
            table_creation = config_settings["export_options"]["allow_table_creation"]
        else:
            table_creation = False
        if load_setting_export_path(logger, config_settings) is None:
            export_path = os.path.join("exports", config_name)
        else:
            export_path = os.path.join(
                load_setting_export_path(logger, config_settings), config_name
            )

        settings = {
            API_TOKEN: load_setting_api_access_token(logger, config_settings),
            SSL_CERT: load_setting_ssl_cert(logger, config_settings),
            SSL_VERIFY: load_setting_ssl_cert(logger, config_settings),
            PROXY_HTTP: load_setting_proxy(logger, config_settings, "http"),
            PROXY_HTTPS: load_setting_proxy(logger, config_settings, "https"),
            EXPORT_PATH: export_path,
            PREFERENCES: load_setting_preference_mapping(logger, config_settings),
            FILENAME_ITEM_ID: get_filename_item_id(logger, config_settings),
            SYNC_DELAY_IN_SECONDS: load_setting_sync_delay(logger, config_settings),
            EXPORT_INACTIVE_ITEMS_TO_CSV: load_export_inactive_items_to_csv(
                logger, config_settings
            ),
            MEDIA_SYNC_OFFSET_IN_SECONDS: load_setting_media_sync_offset(
                logger, config_settings
            ),
            TEMPLATE_IDS: config_settings["export_options"]["template_ids"]
            if config_settings["export_options"]["template_ids"] != ""
            else None,
            SQL_TABLE: config_settings["export_options"]["sql_table"],
            DB_TYPE: config_settings["export_options"]["database_type"],
            DB_USER: config_settings["export_options"]["database_user"],
            DB_PWD: config_settings["export_options"]["database_pwd"],
            DB_SERVER: config_settings["export_options"]["database_server"],
            DB_PORT: config_settings["export_options"]["database_port"],
            DB_NAME: config_settings["export_options"]["database_name"],
            DB_SCHEMA: config_settings["export_options"]["database_schema"],
            USE_REAL_TEMPLATE_NAME: config_settings["export_options"][
                "use_real_template_name"
            ],
            CONFIG_NAME: config_name,
            EXPORT_ARCHIVED: config_settings["export_options"]["export_archived"],
            EXPORT_COMPLETED: config_settings["export_options"]["export_completed"],
            MERGE_ROWS: config_settings["export_options"]["merge_rows"],
            ALLOW_TABLE_CREATION: table_creation,
            ACTIONS_TABLE: load_actions_table(
                config_settings["export_options"]["sql_table"]
            )
            + "_actions",
            ACTIONS_MERGE_ROWS: config_settings["export_options"]["actions_merge_rows"],
        }
    return settings


def configure(logger, path_to_config_file, export_formats, docker_enabled, chunks):
    """
    instantiate and configure logger, load config settings from file, instantiate SafetyCulture SDK
    :param logger:              the logger
    :param path_to_config_file: path to config file
    :param export_formats:      desired export formats
    :return:                    instance of SafetyCulture SDK object, config settings
    """

    config_settings = load_config_settings(logger, path_to_config_file, docker_enabled)
    config_settings[EXPORT_FORMATS] = export_formats
    config_settings["chunks"] = chunks
    if (
        config_settings[PROXY_HTTP] is not None
        and config_settings[PROXY_HTTPS] is not None
    ):
        proxy_settings = {
            "http": config_settings[PROXY_HTTP],
            "https": config_settings[PROXY_HTTPS],
        }
    else:
        proxy_settings = None

    sc_client = sp.SafetyCulture(
        config_settings[API_TOKEN],
        proxy_settings=proxy_settings,
        certificate_settings=config_settings[SSL_CERT],
        ssl_verify=config_settings[SSL_VERIFY],
    )

    if not docker_enabled:
        if config_settings[EXPORT_PATH] is not None:
            if config_settings[CONFIG_NAME] is not None:
                create_directory_if_not_exists(
                    logger, os.path.join(config_settings[EXPORT_PATH])
                )
            else:
                logger.error(
                    "You must set the config_name in your config file before continuing."
                )
                sys.exit()

        else:
            logger.info(
                "No export path was found in "
                + path_to_config_file
                + ", defaulting to /exports"
            )
            config_settings[EXPORT_PATH] = os.path.join(os.getcwd(), "exports")
            if config_settings[CONFIG_NAME] is not None:
                create_directory_if_not_exists(
                    logger, os.path.join(config_settings[EXPORT_PATH])
                )
            else:
                logger.error(
                    "You must set the config_name in your config file before continuing."
                )
                sys.exit()

    return sc_client, config_settings


def rename_config_sample(logger):
    if not os.path.isfile("configs/config.yaml"):
        if os.path.isfile("configs/config.yaml.sample"):
            file_size = os.stat("configs/config.yaml.sample")
            file_size = file_size.st_size
            if file_size <= 666:
                logger.info(
                    'It looks like the config file has not been filled out. Open the folder named "configs" '
                    'and edit the file named "config.yaml.sample" before continuing'
                )
                sys.exit()
            if file_size >= 667:
                logger.info(
                    'It looks like you have not renamed "config.yaml.sample" to "config.yaml". Would you like '
                    "the "
                    "script to do it for you (recommended!)? If you say no, you will need to manually remove "
                    ".sample from the file name.  "
                )
                question = input(
                    "Please type either y (yes) or n (no) and press enter to continue.   "
                )
                if question.startswith("y"):
                    os.rename(r"configs/config.yaml.sample", r"configs/config.yaml")
                else:
                    sys.exit()
            else:
                logger.info(
                    "No config file found. Please either name it config.yaml or specify it with --config."
                )
                sys.exit()
        else:
            logger.info(
                "No config file found. Please either name it config.yaml or specify it with --config."
            )
            sys.exit()


def parse_command_line_arguments(logger):
    """
    Parse command line arguments received, if any
    Print example if invalid arguments are passed

    :param logger:  the logger
    :return:        config_filename passed as argument if any, else DEFAULT_CONFIG_FILENAME
                    export_formats passed as argument if any, else 'pdf'
                    list_preferences if passed as argument, else None
                    do_loop False if passed as argument, else True
    """
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--config", help="config file to use, defaults to " + DEFAULT_CONFIG_FILENAME
    )
    parser.add_argument(
        "--docker",
        nargs="*",
        help="Switches settings to ENV variables for use with docker.",
    )
    parser.add_argument(
        "--format",
        nargs="*",
        help="formats to download, valid options are pdf, "
        "json, docx, csv, media, web-report-link, actions, pickle, sql",
    )
    parser.add_argument(
        "--list_preferences",
        nargs="*",
        help="display all preferences, or restrict to specific"
        " template_id if supplied as additional argument",
    )
    parser.add_argument(
        "--loop", nargs="*", help="execute continuously until interrupted"
    )
    parser.add_argument(
        "--setup", action="store_true", help="Helps set up the script with ease."
    )
    parser.add_argument(
        "--chunks",
        type=int,
        help="Specify a smaller number of chunks when "
        "exporting large numbers of "
        "PDFs, DOCX or Media ",
    )
    args = parser.parse_args()

    if args.setup:
        initial_setup(logger)
        exit()

    if args.docker is None:
        if args.config is None:
            rename_config_sample(logger)

        if args.config is not None:
            config_filename = os.path.join("configs", args.config)
            if os.path.isfile(config_filename):
                config_filename = os.path.join("configs", args.config)
                logger.debug(config_filename + " passed as config argument")
            else:
                logger.error(config_filename + " is either missing or corrupt.")
                sys.exit(1)
        else:
            config_filename = os.path.join("configs", DEFAULT_CONFIG_FILENAME)
    else:
        config_filename = None

    export_formats = ["pdf"]
    if args.format is not None and len(args.format) > 0:
        valid_export_formats = [
            "json",
            "docx",
            "pdf",
            "csv",
            "media",
            "web-report-link",
            "actions",
            "actions-sql",
            "sql",
            "pickle",
            "doc_creation",
        ]
        export_formats = []
        for option in args.format:
            if option not in valid_export_formats:
                print(
                    "{0} is not a valid export format.  Valid options are pdf, json, docx, csv, web-report-link, "
                    "media, actions, pickle, actions_sql, or sql".format(option)
                )
                logger.info("invalid export format argument: {0}".format(option))
            else:
                export_formats.append(option)

    chunks = args.chunks if args.chunks is not None else 100
    loop_enabled = True if args.loop is not None else False
    docker_enabled = True if args.docker is not None else False

    return (
        config_filename,
        export_formats,
        args.list_preferences,
        loop_enabled,
        docker_enabled,
        chunks,
    )


def get_port(db_type):
    if db_type == "mssql":
        return "1433"
    elif db_type == "postgres":
        return "5432"
    elif db_type == "mysql":
        return "3308"
    else:
        return ""


def get_schema(db_type):
    if db_type == "mssql":
        return "dbo"
    elif db_type == "postgres":
        return "public"
    elif db_type == "mysql":
        return ""
    else:
        return ""


def get_default_db(db_type):
    if db_type == "sql":
        return "iAuditor?driver=ODBC Driver 17 for SQL Server"
    else:
        return "auditor"


def ask_question(logger, str, q_type, choices=None, special=None, default=""):
    response = None
    if q_type == "multi":
        if default == "":
            default = []
        response = questionary.select(str, default=default, choices=choices).ask()
    elif q_type == "text":
        response = questionary.text(str, default=default).ask()

    elif q_type == "int":
        response = questionary.text(str, default=default).ask()
        if response:
            response = int(response)
        else:
            response = default

    elif q_type == "bool":
        response = questionary.confirm(str).ask()

    elif q_type == "password":
        response = questionary.password(str).ask()

    if special == "config_name":
        response = sanitise_config_name(logger, response)

    return response


def sanitise_config_name(logger, config_name):
    if config_name is None:
        logger.info("The Config Name has been left blank, defaulting to iauditor.")
        config_name = "iauditor"
    elif " " in config_name:
        config_name = config_name.replace(" ", "_")
    else:
        config_name = config_name
    if re.match("^[A-Za-z0-9_-]*$", config_name):
        config_name = config_name

    return config_name


def setup_database(logger):
    db_type = questionary.select(
        "Which database type are you using?",
        choices=[
            {"name": "SQL", "value": "mssql"},
            {"name": "MySQL", "value": "mysql"},
            {"name": "Postgres", "value": "postgres"},
        ],
    ).ask()
    db_settings = {
        "database_type": db_type,
        "database_user": questionary.text("Database username").ask(),
        "database_pwd": questionary.password("Database Password").ask(),
        "database_server": questionary.text("Database Server Address").ask(),
        "database_port": ask_question(
            logger, "Database Port", "int", default=get_port(db_type)
        ),
        "database_schema": questionary.text(
            "Database Schema", default=get_schema(db_type)
        ).ask(),
        "database_name": questionary.text(
            "Database Name", default=get_default_db(db_type)
        ).ask(),
    }

    if db_type == "mssql":
        db_settings["database_name"] = (
            db_settings["database_name"] + "?driver=ODBC Driver 17 for SQL Server"
        )

    return db_settings


def box_print(h):
    if h:
        return Panel(h, box=box.HEAVY)
    else:
        return ""


def update_key(logger, settings, key):
    key_to_update = questions.get(key)
    if key_to_update:
        q = key_to_update["question"]
        t = key_to_update["type"]
        parent = key_to_update["parent"]
        if t == "api_token":
            new_value = sp.interactive_login()
            if not new_value:
                new_value = sp.interactive_login()
        else:
            if "options" in key_to_update:
                choices = key_to_update.get("options")
            else:
                choices = None
            if parent:
                current_value = settings[parent].get(key)
            else:
                current_value = settings.get(key)
            h = f"""
                Currently set to: {current_value} 
                {key_to_update['header']}
                """
            print(box_print(h))
            new_value = ask_question(logger, q, t, choices, special=key_to_update)
        if key_to_update["parent"]:
            settings[parent][key] = new_value
        else:
            settings[key] = new_value
    # elif key.startswith(("database_", "sql_")):
    #     db_settings = setup_database(logger)
    #     print(db_settings)
    #     sql_test = test_sql_settings(logger, db_settings)
    #     if sql_test:
    #         logger.info("Connected successfully, writing settings to config file.")
    #         for k, v in db_settings.items():
    #             settings["export_options"][k] = v
    #     else:
    #         logger.warning("Connection unsuccessful, trying again.")
    #         setup_database(logger)
    else:
        print(f"{key} is not a recognised setting. If unsure, start a new config file.")


def modify_choice(logger, settings, config_path, current_ls):
    options = []
    for k, v in settings.items():
        options.append(Separator())
        if k == "config_name":
            options.append(f"{k} - Current setting: {v}")
        else:
            for inner_k, inner_v in v.items():
                if inner_k == "sql_table":
                    options.append(Separator())
                options.append(f"{inner_k} - Current setting: {inner_v}")
    exit_no_save = "Exit without saving changes"
    exit_save = "Exit and save"
    test_sql = "Test database settings"
    change_ls = "Change time to search from"
    change_ls_opt = f"{change_ls} - Current setting: {current_ls}"
    if exit_no_save not in options:
        exit_list = [exit_save, exit_no_save, test_sql, change_ls_opt, Separator()]
        options = exit_list + options
    to_modify = questionary.select(
        "Select an option to modify:", choices=options
    ).ask()
    if to_modify:
        to_modify = to_modify.split("-")[0].strip()
    else:
        sys.exit()
    if to_modify == exit_save or to_modify == test_sql or to_modify == change_ls:
        if settings[CONFIG_NAME] is not None:
            config_name = settings[CONFIG_NAME]
        else:
            config_name = "iauditor"
        with open(config_path, "w") as file:
            updated_config = yaml.dump(
                settings, default_flow_style=False, sort_keys=False
            )
            updated_config = updated_config.replace("-", "")
            updated_config = updated_config.replace("null", "")
            file.write(updated_config.replace("-", ""))
        if to_modify == test_sql:
            sql_test_result = test_sql_settings(logger, settings)
            if not sql_test_result:
                modify_choice(logger, settings, config_path, current_ls)
        elif to_modify == change_ls:
            parsed_date = parse_ls(current_ls)
            if parsed_date:
                pretty_file_name = parsed_date.strftime("%Y-%m-%d at %H%M")
                print(
                    box_print(
                        f"""
                We are currently searching for inspections completed after {pretty_file_name}.
                You can use natural language here. For example if you wanted to start the search from March 22nd 2018, 
                you could simply type 'March 22nd 2018'
                """
                    )
                )
                change_ls = ask_question(
                    logger,
                    "When should we search from? If you want to start from the beginning, "
                    "just press enter.",
                    "text",
                )
                new_date = parse_ls(change_ls)
                if new_date:
                    new_date = new_date.strftime("%Y-%m-%dT%H:%M:%S%z")
                    update_sync_marker_file(str(new_date), config_name)
                modify_choice(logger, settings, config_path, new_date)
            else:
                logger.info("All done. Re run the tool to use your new configuration. ")
                sys.exit()
    elif to_modify == exit_no_save:
        logger.info("Changes discarded.")
        sys.exit()
    elif not to_modify:
        modify_choice(logger, settings, config_path, current_ls)
    else:
        update_key(logger, settings, to_modify)
        modify_choice(logger, settings, config_path, current_ls)


def initial_setup(logger):
    """
    Creates a new directory in current working directory called 'iauditor_exports_folder'.  If 'iauditor_exports_folder'
    already exists the setup script will notify user that the folder exists and exit. Default config file placed
    in directory, with user API Token. User is asked for iAuditor credentials in order to generate their
    API token.
    :param logger:  the logger
    """

    # setup variables
    current_directory_path = os.getcwd()
    box_print("iAuditor Export Tool Setup")
    exports_folder_name = "configs"

    create_directory_if_not_exists(logger, exports_folder_name)

    config_files = os.listdir(exports_folder_name)

    if config_files:
        if len(config_files) > 1:
            config_to_edit = questionary.select(
                "It looks like you already have some config files. Which one do you want to modify?",
                choices=config_files,
            ).ask()
        else:
            config_to_edit = config_files[0]
        config_path = os.path.join(exports_folder_name, config_to_edit)
        with open(os.path.join(exports_folder_name, config_to_edit)) as file:
            settings = yaml.load(file, Loader=yaml.FullLoader)
    else:
        settings = model_config
        config_to_edit = "config.yaml"
        config_path = os.path.join(exports_folder_name, config_to_edit)
    if "config_name" in settings:
        config_name = settings["config_name"]
    else:
        print(
            "Your config file is missing config_name. If you delete your config file and "
            "re run the tool with --setup it will be recreated correctly."
        )
        sys.exit()

    current_ls = get_last_successful(logger, config_name)
    # parsed_date = parse_ls(current_ls)
    # if parsed_date:
    #     pretty_file_name = parsed_date.strftime("%Y-%m-%d at %H%M")
    #
    #     update_ls_q = ask_question(
    #         logger,
    #         f"We are currently searching for inspections completed after {pretty_file_name}, would you "
    #         f"like to change this?",
    #         "bool",
    #     )
    #     if update_ls_q:
    #         print(
    #             box_print(
    #                 """
    #         You can use natural language here. For example if you wanted to start the search from March 22nd 2018,
    #         you could simply type 'March 22nd 2018'
    #         """
    #             )
    #         )
    #         change_ls = ask_question(
    #             logger,
    #             "When should we search from? If you want to start from the beginning, "
    #             "just press enter.",
    #             "text",
    #         )
    #         new_date = parse_ls(change_ls)
    #         if new_date:
    #             new_date = new_date.strftime("%Y-%m-%dT%H:%M:%S%z")
    #             update_sync_marker_file(str(new_date), config_name)

    modify_choice(logger, settings, config_path, current_ls)

    sys.exit()
