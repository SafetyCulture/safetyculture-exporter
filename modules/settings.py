import argparse
import re
import yaml

from modules.global_variables import *
from modules.logger import log_critical_error, create_directory_if_not_exists
from safetypy import safetypy as sp


def load_setting_api_access_token(logger, config_settings):
    """
    Attempt to parse API token from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 API token if valid, else None
    """
    try:
        api_token = config_settings['API']['token']
        token_is_valid = re.match('^[a-f0-9]{64}$', api_token)
        if token_is_valid:
            logger.debug('API token matched expected pattern')
            return api_token
        else:
            logger.error('API token failed to match expected pattern')
            return None
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception parsing API token from config.yaml')
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
        token_is_valid = re.match('^[a-f0-9]{64}$', api_token)
        if token_is_valid:
            logger.debug('API token matched expected pattern')
            return api_token
        else:
            logger.error('API token failed to match expected pattern')
            return None
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception parsing API token from config.yaml')
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
        if config_settings['export_options']['merge_rows'] is True:
            logger.info('Merge rows is enabled, turning on the export of inactive items.')
            export_inactive_items_to_csv = True
        else:
            export_inactive_items_to_csv = config_settings['export_options']['export_inactive_items']
            if not isinstance(export_inactive_items_to_csv, bool):
                logger.info('Invalid export_inactive_items value from configuration file, defaulting to true')
                export_inactive_items_to_csv = DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV
        return export_inactive_items_to_csv
    except Exception as ex:
        log_critical_error(logger, ex,
                           'Exception parsing export_inactive_items from the configuration file, defaulting to {0}'.
                           format(str(DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV)))
        return DEFAULT_EXPORT_INACTIVE_ITEMS_TO_CSV


def load_setting_sync_delay(logger, config_settings):
    """
    Attempt to parse delay between sync loops from config settings

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 extracted sync delay if valid, else DEFAULT_SYNC_DELAY_IN_SECONDS
    """
    try:
        sync_delay = config_settings['export_options']['sync_delay_in_seconds']
        sync_delay_is_valid = re.match('^[0-9]+$', str(sync_delay))
        if sync_delay_is_valid and sync_delay >= 0:
            if sync_delay < DEFAULT_SYNC_DELAY_IN_SECONDS:
                '{0} seconds'.format(logger.info(
                    'Sync delay is less than the minimum recommended value of ' + str(DEFAULT_SYNC_DELAY_IN_SECONDS)))
            return sync_delay
        else:
            logger.info('Invalid sync_delay_in_seconds from the configuration file, defaulting to {0}'.format(str(
                DEFAULT_SYNC_DELAY_IN_SECONDS)))
            return DEFAULT_SYNC_DELAY_IN_SECONDS
    except Exception as ex:
        log_critical_error(logger, ex,
                           'Exception parsing sync_delay from the configuration file, defaulting to {0}'.format(str(
                               DEFAULT_SYNC_DELAY_IN_SECONDS)))
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
        preference_settings = config_settings['export_options']['preferences']
        if preference_settings is not None:
            preference_lines = preference_settings.split(' ')
            for preference in preference_lines:
                template_id = preference[:preference.index(':')]
                if template_id not in preference_mapping.keys():
                    preference_mapping[template_id] = preference
        return preference_mapping
    except KeyError:
        logger.debug('No preference key in the configuration file')
        return None
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception getting preferences from the configuration file')
        return None


def load_setting_export_path(logger, config_settings):
    """
    Attempt to extract export path from config settings

    :param config_settings:  config settings loaded from config file
    :param logger:           the logger
    :return:                 export path, None if path is invalid or missing
    """
    try:
        export_path = config_settings['export_options']['export_path']
        if export_path is not None:
            return export_path
        else:
            return None
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception getting export path from the configuration file')
        return None


def load_setting_media_sync_offset(logger, config_settings):
    """

    :param logger:           the logger
    :param config_settings:  config settings loaded from config file
    :return:                 media sync offset parsed from file, else default media sync offset
                             defined as global constant
    """
    try:
        media_sync_offset = config_settings['export_options']['media_sync_offset_in_seconds']
        if media_sync_offset is None or media_sync_offset < 0 or not isinstance(media_sync_offset, int):
            media_sync_offset = DEFAULT_MEDIA_SYNC_OFFSET_IN_SECONDS
        return media_sync_offset
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception parsing media sync offset from config file')
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
    if filename_item_id == AUDIT_TITLE_ITEM_ID and 'audit_data' in audit_json.keys() \
            and 'name' in audit_json['audit_data'].keys():
        return audit_json['audit_data']['name'].replace('/', '_')
    for item in audit_json['header_items']:
        if item['item_id'] == filename_item_id:
            if 'responses' in item.keys():
                if 'text' in item['responses'].keys() and item['responses']['text'].strip() != '':
                    return item['responses']['text']
    return None


def get_filename_item_id(logger, config_settings):
    """
    Attempt to parse item_id for file naming from config settings

    :param logger:          the logger
    :param config_settings: config settings loaded from config file
    :return:                item_id extracted from config_settings if valid, else None
    """
    try:
        filename_item_id = config_settings['export_options']['filename']
        if filename_item_id is not None:
            return filename_item_id
        else:
            return None
    except Exception as ex:
        log_critical_error(logger, ex, 'Exception retrieving setting "filename" from the configuration file')
        return None


def set_env_defaults(name, env_var, logger):
    # if env_var is None or '':
    if not env_var:
        if name == 'CONFIG_NAME':
            logger.error('You must set the CONFIG_NAME')
            sys.exit()
        if name == 'DB_SCHEMA':
            env_var = 'dbo'
        if name.startswith('DB_'):
            env_var = None
        if name == 'SQL_TABLE':
            env_var = None
        if name == 'TEMPLATE_IDS':
            env_var = None
        else:
            env_var = 'false'
    print(name, ' set to ', env_var)
    return env_var


def load_setting_ssl_cert(logger, config_settings):
    cert_location = None
    if 'ssl_cert' in config_settings['API']:
        if config_settings['API']['ssl_cert']:
            cert_location = config_settings['API']['ssl_cert']
    return cert_location


def load_setting_proxy(logger, config_settings, http_or_https):
    proxy = None
    if http_or_https == 'https':
        if 'proxy_https' in config_settings['API']:
            if config_settings['API']['proxy_https']:
                proxy = config_settings['API']['proxy_https']
    elif http_or_https == 'http':
        if 'proxy_http' in config_settings['API']:
            if config_settings['API']['proxy_http']:
                proxy = config_settings['API']['proxy_http']
    else:
        proxy = None
    return proxy


def load_actions_table(actions_table_name):
    if actions_table_name is None:
        actions_table_name = 'iauditor'
        return actions_table_name
    else:
        return actions_table_name


def load_config_settings(logger, path_to_config_file, docker_enabled):
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
            API_TOKEN: docker_load_setting_api_access_token(logger, os.environ['API_TOKEN']),
            EXPORT_PATH: None,
            # PREFERENCES: load_setting_preference_mapping(logger, config_settings),
            # FILENAME_ITEM_ID: get_filename_item_id(logger, config_settings),
            SYNC_DELAY_IN_SECONDS: int(os.environ['SYNC_DELAY_IN_SECONDS']),
            # EXPORT_INACTIVE_ITEMS_TO_CSV: load_export_inactive_items_to_csv(logger, config_settings),
            MEDIA_SYNC_OFFSET_IN_SECONDS: int(os.environ['MEDIA_SYNC_OFFSET_IN_SECONDS']),
            TEMPLATE_IDS: set_env_defaults('TEMPLATE_IDS', os.environ['TEMPLATE_IDS'], logger),
            SQL_TABLE: set_env_defaults('SQL_TABLE', os.environ['SQL_TABLE'], logger),
            DB_TYPE: set_env_defaults('DB_TYPE', os.environ['DB_TYPE'], logger),
            DB_USER: set_env_defaults('DB_USER', os.environ['DB_USER'], logger),
            DB_PWD: set_env_defaults('DB_PWD', os.environ['DB_PWD'], logger),
            DB_SERVER: set_env_defaults('DB_SERVER', os.environ['DB_SERVER'], logger),
            DB_PORT: set_env_defaults('DB_PORT', os.environ['DB_PORT'], logger),
            DB_NAME: set_env_defaults('DB_NAME', os.environ['DB_NAME'], logger),
            DB_SCHEMA: set_env_defaults('DB_SCHEMA', os.environ['DB_SCHEMA'], logger),
            USE_REAL_TEMPLATE_NAME: set_env_defaults('USE_REAL_TEMPLATE_NAME', os.environ['USE_REAL_TEMPLATE_NAME'],
                                                     logger),
            CONFIG_NAME: set_env_defaults('CONFIG_NAME', os.environ['CONFIG_NAME'], logger),
            EXPORT_ARCHIVED: set_env_defaults('EXPORT_ARCHIVED', os.environ['EXPORT_ARCHIVED'], logger),
            EXPORT_COMPLETED: set_env_defaults('EXPORT_COMPLETED', os.environ['EXPORT_COMPLETED'], logger),
            MERGE_ROWS: set_env_defaults('MERGE_ROWS', os.environ['MERGE_ROWS'], logger),
            ALLOW_TABLE_CREATION: set_env_defaults('ALLOW_TABLE_CREATION', os.environ['ALLOW_TABLE_CREATION'], logger),
            ACTIONS_TABLE: 'iauditor_actions_data',
            ACTIONS_MERGE_ROWS: set_env_defaults('ACTIONS_MERGE_ROWS', os.environ['ACTIONS_MERGE_ROWS'], logger),
            PREFERENCES: None,
            FILENAME_ITEM_ID: None,
            EXPORT_INACTIVE_ITEMS_TO_CSV: None
        }
    else:
        config_settings = yaml.safe_load(open(path_to_config_file))
        if config_settings['config_name'] is None:
            logger.info('The Config Name has been left blank, defaulting to iauditor.')
            config_name = 'iauditor'
        elif ' ' in config_settings['config_name']:
            config_name = config_settings['config_name'].replace(' ', '_')
        else:
            config_name = config_settings['config_name']

        if re.match("^[A-Za-z0-9_-]*$", config_name):
            config_name = config_name
        else:
            logger.critical('Config name can only contain letters, numbers, hyphens or underscores.')
            sys.exit()
        if 'allow_table_creation' in config_settings['export_options']:
            table_creation = config_settings['export_options']['allow_table_creation']
        else:
            table_creation = False
        if load_setting_export_path(logger, config_settings) is None:
            export_path = os.path.join('exports', config_name)
        else:
            export_path = os.path.join(load_setting_export_path(logger, config_settings), config_name)

        settings = {
            API_TOKEN: load_setting_api_access_token(logger, config_settings),
            SSL_CERT: load_setting_ssl_cert(logger, config_settings),
            PROXY_HTTP: load_setting_proxy(logger, config_settings, 'http'),
            PROXY_HTTPS: load_setting_proxy(logger, config_settings, 'https'),
            EXPORT_PATH: export_path,
            PREFERENCES: load_setting_preference_mapping(logger, config_settings),
            FILENAME_ITEM_ID: get_filename_item_id(logger, config_settings),
            SYNC_DELAY_IN_SECONDS: load_setting_sync_delay(logger, config_settings),
            EXPORT_INACTIVE_ITEMS_TO_CSV: load_export_inactive_items_to_csv(logger, config_settings),
            MEDIA_SYNC_OFFSET_IN_SECONDS: load_setting_media_sync_offset(logger, config_settings),
            TEMPLATE_IDS: config_settings['export_options']['template_ids'],
            SQL_TABLE: config_settings['export_options']['sql_table'],
            DB_TYPE: config_settings['export_options']['database_type'],
            DB_USER: config_settings['export_options']['database_user'],
            DB_PWD: config_settings['export_options']['database_pwd'],
            DB_SERVER: config_settings['export_options']['database_server'],
            DB_PORT: config_settings['export_options']['database_port'],
            DB_NAME: config_settings['export_options']['database_name'],
            DB_SCHEMA: config_settings['export_options']['database_schema'],
            USE_REAL_TEMPLATE_NAME: config_settings['export_options']['use_real_template_name'],
            CONFIG_NAME: config_name,
            EXPORT_ARCHIVED: config_settings['export_options']['export_archived'],
            EXPORT_COMPLETED: config_settings['export_options']['export_completed'],
            MERGE_ROWS: config_settings['export_options']['merge_rows'],
            ALLOW_TABLE_CREATION: table_creation,
            ACTIONS_TABLE: load_actions_table(config_settings['export_options']['sql_table']) + '_actions',
            ACTIONS_MERGE_ROWS: config_settings['export_options']['actions_merge_rows']
        }
    return settings


def configure(logger, path_to_config_file, export_formats, docker_enabled):
    """
    instantiate and configure logger, load config settings from file, instantiate SafetyCulture SDK
    :param logger:              the logger
    :param path_to_config_file: path to config file
    :param export_formats:      desired export formats
    :return:                    instance of SafetyCulture SDK object, config settings
    """

    config_settings = load_config_settings(logger, path_to_config_file, docker_enabled)
    config_settings[EXPORT_FORMATS] = export_formats
    if config_settings[PROXY_HTTP] is not None and config_settings[PROXY_HTTPS] is not None:
        proxy_settings = {
            "http": config_settings[PROXY_HTTP],
            "https": config_settings[PROXY_HTTPS]
        }
    else:
        proxy_settings = None

    sc_client = sp.SafetyCulture(config_settings[API_TOKEN])

    if config_settings[EXPORT_PATH] is not None:
        if config_settings[CONFIG_NAME] is not None:
            create_directory_if_not_exists(logger, os.path.join(config_settings[EXPORT_PATH]))
        else:
            logger.error("You must set the config_name in your config file before continuing.")
            sys.exit()
    else:
        logger.info('No export path was found in ' + path_to_config_file + ', defaulting to /exports')
        config_settings[EXPORT_PATH] = os.path.join(os.getcwd(), 'exports')
        if config_settings[CONFIG_NAME] is not None:
            create_directory_if_not_exists(logger, os.path.join(config_settings[EXPORT_PATH]))
        else:
            logger.error("You must set the config_name in your config file before continuing.")
            sys.exit()

    return sc_client, config_settings


def rename_config_sample(logger):
    if os.path.isfile('configs/config.yaml.sample'):
        file_size = os.stat('configs/config.yaml.sample')
        file_size = file_size.st_size
        print(file_size)
        if file_size <= 667:
            logger.info('It looks like the config file has not been filled out. Open the folder named "configs" '
                        'and edit the file named "config.yaml.sample" before continuing')
            sys.exit()
        if file_size > 667:
            logger.info('It looks like you have edited config.yaml.sample but have not renamed it. Would you like '
                        'the '
                        'script to do it for you (recommended!)? If you say no, you will need to manually remove '
                        '.sample from the file name.')
            question = input('     (y/n)')
            if question == 'y':
                os.rename(r'configs/config.yaml.sample', r'configs/config.yaml')
            else:
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
    parser.add_argument('--config', help='config file to use, defaults to ' + DEFAULT_CONFIG_FILENAME)
    parser.add_argument('--docker', nargs='*', help='Switches settings to ENV variables for use with docker.')
    parser.add_argument('--format', nargs='*', help='formats to download, valid options are pdf, '
                                                    'json, docx, csv, media, web-report-link, actions, pickle, sql')
    parser.add_argument('--list_preferences', nargs='*', help='display all preferences, or restrict to specific'
                                                              ' template_id if supplied as additional argument')
    parser.add_argument('--loop', nargs='*', help='execute continuously until interrupted')
    parser.add_argument('--setup', action='store_true', help='Automatically create new directory containing the '
                                                             'necessary config file.'
                                                             'Directory will be named iAuditor Audit Exports, and will '
                                                             'be placed in your current directory')
    args = parser.parse_args()

    if args.config is None:
        rename_config_sample(logger)

    if args.config is not None:
        config_filename = os.path.join('configs', args.config)
        print(args.config)
        if os.path.isfile(config_filename):
            config_filename = os.path.join('configs', args.config)
            logger.debug(config_filename + ' passed as config argument')
        else:
            logger.error(config_filename + ' is either missing or corrupt.')
            rename_config_sample(logger)
            sys.exit(1)
    else:
        config_filename = os.path.join('configs', DEFAULT_CONFIG_FILENAME)

    export_formats = ['pdf']
    if args.format is not None and len(args.format) > 0:
        valid_export_formats = ['json', 'docx', 'pdf', 'csv', 'media', 'web-report-link', 'actions', 'actions-sql',
                                'sql', 'pickle', 'doc_creation']
        export_formats = []
        for option in args.format:
            if option not in valid_export_formats:
                print('{0} is not a valid export format.  Valid options are pdf, json, docx, csv, web-report-link, '
                      'media, actions, pickle, actions_sql, or sql'.format(option))
                logger.info('invalid export format argument: {0}'.format(option))
            else:
                export_formats.append(option)

    loop_enabled = True if args.loop is not None else False
    docker_enabled = True if args.docker is not None else False

    return config_filename, export_formats, args.list_preferences, loop_enabled, docker_enabled

