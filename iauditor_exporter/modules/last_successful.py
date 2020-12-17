import os
import sys

import dateparser
from dateparser.search import search_dates

from iauditor_exporter.modules.global_variables import (
    ACTIONS_SYNC_MARKER_FILENAME,
    SYNC_MARKER_FILENAME,
)
from iauditor_exporter.modules.logger import (
    log_critical_error,
    create_directory_if_not_exists,
)


def parse_ls(date):
    parsed_date = dateparser.parse(date)
    if not parsed_date:
        date = search_dates(date)
    else:
        date = parsed_date
    if not date:
        print(
            "The date provided cannot be parsed. Defaulting to the beginning of your account."
        )
        date = dateparser.parse("2000-01-01T00:00:00.000Z")
    return date


def set_last_successful_file_name(config_name, audit_or_action):
    if config_name is not None and audit_or_action == "audits":
        last_successful_file = "last_successful/last_successful-{}.txt".format(
            config_name
        )
    elif config_name is not None and audit_or_action == "actions":
        last_successful_file = "last_successful/last_successful_actions_export-{}.txt".format(
            config_name
        )
    elif audit_or_action == "audits":
        last_successful_file = SYNC_MARKER_FILENAME
    elif audit_or_action == "actions":
        last_successful_file = ACTIONS_SYNC_MARKER_FILENAME
    else:
        sys.exit()
    return last_successful_file


def update_sync_marker_file(date_modified, config_name):
    """
    Replaces the contents of the sync marker file with the most
    recent modified_at date time value from audit JSON data

    :param date_modified:   modified_at value from most recently downloaded audit JSON
    :return:
    """

    last_successful_file = set_last_successful_file_name(config_name, "audits")

    with open(last_successful_file, "w") as sync_marker_file:
        sync_marker_file.write(date_modified)


def get_last_successful(logger, config_name):
    """
    Read the date and time of the last successfully exported audit data from the sync marker file

    :param logger:  the logger
    :return:        A datetime value (or 2000-01-01 if syncing since the 'beginning of time')
    :config_name:

    """
    last_successful_file = set_last_successful_file_name(config_name, "audits")

    if os.path.isfile(last_successful_file):
        with open(last_successful_file, "r+") as last_run:
            check_for_rows = last_run.readlines()
            if check_for_rows:
                last_successful = check_for_rows[0]
                last_successful = last_successful.strip()
            else:
                last_successful = "2000-01-01T00:00:00.000Z"

    else:
        beginning_of_time = "2000-01-01T00:00:00.000Z"
        last_successful = beginning_of_time
        create_directory_if_not_exists(logger, "last_successful")
        with open(last_successful_file, "w") as last_run:
            last_run.write(last_successful)
        logger.info(
            "Searching for audits since the beginning of time: " + beginning_of_time
        )
    return last_successful


def update_actions_sync_marker_file(logger, date_modified, config_name):
    """
    Replaces the contents of the actions sync marker file with the the date/time string provided
    :param logger:   The logger
    :param date_modified:   ISO string

    """

    last_successful_file = set_last_successful_file_name(config_name, "actions")

    try:
        with open(last_successful_file, "w") as actions_sync_marker_file:
            actions_sync_marker_file.write(date_modified)
    except Exception as ex:
        log_critical_error(
            logger, ex, "Unable to open " + last_successful_file + " for writing"
        )
        exit()


def get_last_successful_actions_export(logger, config_name):
    """
    Reads the actions sync marker file to determine the date and time of the most last successfully exported action.
    The actions sync marker file is expected to contain a single ISO formatted datetime string.
    :param logger:  the logger
    :return:        A datetime value (or 2000-01-01 if syncing since the 'beginning of time')
    """

    last_successful_file = set_last_successful_file_name(config_name, "actions")

    if os.path.isfile(last_successful_file):
        with open(last_successful_file, "r+") as last_run:
            last_successful_actions_export = last_run.readlines()[0].strip()
            logger.info(
                "Searching for actions modified after " + last_successful_actions_export
            )
    else:
        beginning_of_time = "2000-01-01T00:00:00.000Z"
        last_successful_actions_export = beginning_of_time
        with open(last_successful_file, "w") as last_run:
            last_run.write(last_successful_actions_export)
        logger.info(
            "Searching for actions since the beginning of time: " + beginning_of_time
        )
    return last_successful_actions_export
