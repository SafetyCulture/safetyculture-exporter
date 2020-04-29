def update_sync_marker_file(date_modified):
    """
    Replaces the contents of the sync marker file with the most
    recent modified_at date time value from audit JSON data

    :param date_modified:   modified_at value from most recently downloaded audit JSON
    :return:
    """
    with open(SYNC_MARKER_FILENAME, 'w') as sync_marker_file:
        sync_marker_file.write(date_modified)


def get_last_successful(logger):
    """
    Read the date and time of the last successfully exported audit data from the sync marker file

    :param logger:  the logger
    :return:        A datetime value (or 2000-01-01 if syncing since the 'beginning of time')
    """
    if os.path.isfile(SYNC_MARKER_FILENAME):
        with open(SYNC_MARKER_FILENAME, 'r+') as last_run:
            last_successful = last_run.readlines()[0]
            last_successful = last_successful.strip()

    else:
        beginning_of_time = '2000-01-01T00:00:00.000Z'
        last_successful = beginning_of_time
        create_directory_if_not_exists(logger, 'last_successful')
        with open(SYNC_MARKER_FILENAME, 'w') as last_run:
            last_run.write(last_successful)
        logger.info('Searching for audits since the beginning of time: ' + beginning_of_time)
    return last_successful


def update_actions_sync_marker_file(logger, date_modified):
    """
    Replaces the contents of the actions sync marker file with the the date/time string provided
    :param logger:   The logger
    :param date_modified:   ISO string
    """
    try:
        with open(ACTIONS_SYNC_MARKER_FILENAME, 'w') as actions_sync_marker_file:
            actions_sync_marker_file.write(date_modified)
    except Exception as ex:
        log_critical_error(logger, ex, 'Unable to open ' + ACTIONS_SYNC_MARKER_FILENAME + ' for writing')
        exit()


def get_last_successful_actions_export(logger):
    """
    Reads the actions sync marker file to determine the date and time of the most last successfully exported action.
    The actions sync marker file is expected to contain a single ISO formatted datetime string.
    :param logger:  the logger
    :return:        A datetime value (or 2000-01-01 if syncing since the 'beginning of time')
    """
    if os.path.isfile(ACTIONS_SYNC_MARKER_FILENAME):
        with open(ACTIONS_SYNC_MARKER_FILENAME, 'r+') as last_run:
            last_successful_actions_export = last_run.readlines()[0]
            logger.info('Searching for actions modified after ' + last_successful_actions_export)
    else:
        beginning_of_time = '2000-01-01T00:00:00.000Z'
        last_successful_actions_export = beginning_of_time
        with open(ACTIONS_SYNC_MARKER_FILENAME, 'w') as last_run:
            last_run.write(last_successful_actions_export)
        logger.info('Searching for actions since the beginning of time: ' + beginning_of_time)
    return last_successful_actions_export