import json
import os
from datetime import datetime

import unicodecsv as csv

import iauditor_exporter.modules.csvExporter as csvExporter
from iauditor_exporter.modules.actions import transform_action_object_to_list
from iauditor_exporter.modules.global_variables import (
    EXPORT_PATH,
    EXPORT_INACTIVE_ITEMS_TO_CSV,
    USE_REAL_TEMPLATE_NAME,
    CONFIG_NAME,
    ACTIONS_EXPORT_FILENAME,
)
from iauditor_exporter.modules.last_successful import (
    get_last_successful_actions_export,
    update_actions_sync_marker_file,
)
from iauditor_exporter.modules.logger import (
    log_critical_error,
    create_directory_if_not_exists,
)
from iauditor_exporter.modules.sql import save_exported_actions_to_db


def save_exported_document(logger, export_dir, export_doc, filename, extension):
    """
    Write exported document to disk at specified location with specified file name.
    Any existing file with the same name will be overwritten.
    :param logger:      the logger
    :param export_dir:  path to directory for exports
    :param export_doc:  export document to write
    :param filename:    filename to give exported document
    :param extension:   extension to give exported document
    """
    path = os.path.join(export_dir, extension)
    file_path = os.path.join(export_dir, extension, filename + "." + extension)
    create_directory_if_not_exists(logger, path)
    if os.path.isfile(file_path):
        logger.info("Overwriting existing report at " + file_path)
    try:
        with open(file_path, "wb") as export_file:
            export_file.write(export_doc)
    except Exception as ex:
        log_critical_error(
            logger, ex, "Exception while writing" + file_path + " to file"
        )


def export_audit_pdf_word(
    logger, sc_client, settings, audit_id, preference_id, export_format, export_filename
):
    """
    Save Audit to disk in PDF or MS Word format
    :param logger:      The logger
    :param sc_client:   instance of safetypy.SafetyCulture class
    :param settings:    Settings from command line and configuration file
    :param audit_id:    Unique audit UUID
    :param preference_id:   Unique preference UUID
    :param export_format:       'pdf' or 'docx' string
    :param export_filename:     String indicating what to name the exported audit file
    """
    export_doc = sc_client.get_export(audit_id, preference_id, export_format)
    save_exported_document(
        logger, settings[EXPORT_PATH], export_doc, export_filename, export_format
    )


def export_audit_json(logger, settings, audit_json, export_filename):
    """
    Save audit JSON to disk
    :param logger:      The logger
    :param settings:    Settings from the command line and configuration file
    :param audit_json:  Audit JSON
    :param export_filename:     String indicating what to name the exported audit file
    """
    export_format = "json"
    export_doc = json.dumps(audit_json, indent=4)
    save_exported_document(
        logger,
        settings[EXPORT_PATH],
        export_doc.encode(),
        export_filename,
        export_format,
    )


def export_audit_csv(settings, audit_json):
    """
    Save audit CSV to disk.
    :param settings:    Settings from command line and configuration file
    :param audit_json:  Audit JSON
    """

    csv_exporter = csvExporter.CsvExporter(
        audit_json, settings[EXPORT_INACTIVE_ITEMS_TO_CSV]
    )
    count = 0
    if settings[USE_REAL_TEMPLATE_NAME] is False:
        csv_export_filename = audit_json["template_id"]
    elif settings[USE_REAL_TEMPLATE_NAME] is True:
        csv_export_filename = (
            audit_json["template_data"]["metadata"]["name"]
            + " - "
            + audit_json["template_id"]
        )
        csv_export_filename = csv_export_filename.replace("/", " ").replace("\\", " ")
    elif settings[USE_REAL_TEMPLATE_NAME] is str and settings[
        USE_REAL_TEMPLATE_NAME
    ].startswith("single_file"):
        csv_export_filename = settings[CONFIG_NAME]
    else:
        csv_export_filename = audit_json["template_id"]

    for row in csv_exporter.audit_table:
        count += 1
        row[0] = count

    csv_exporter.append_converted_audit_to_bulk_export_file(
        os.path.join(settings[EXPORT_PATH], csv_export_filename + ".csv")
    )


def save_exported_actions_to_csv_file(logger, export_path, actions_array, config_name):
    """
    Write Actions to 'iauditor_actions.csv' on disk at specified location
    :param logger:          the logger
    :param export_path:     path to directory for exports
    :param actions_array:   Array of action objects to be converted to CSV and saved to disk
    """
    if not actions_array:
        logger.info(
            "No actions returned after "
            + get_last_successful_actions_export(logger, config_name)
        )
        return
    filename = ACTIONS_EXPORT_FILENAME
    file_path = os.path.join(export_path, filename)
    logger.info("Exporting " + str(len(actions_array)) + " actions to " + file_path)
    if os.path.isfile(file_path):
        actions_csv = open(file_path, "ab")
        actions_csv_wr = csv.writer(actions_csv, dialect="excel", quoting=csv.QUOTE_ALL)
    else:
        actions_csv = open(file_path, "wb")
        actions_csv_wr = csv.writer(actions_csv, dialect="excel", quoting=csv.QUOTE_ALL)
        actions_csv_wr.writerow(
            [
                "actionId",
                "title",
                "description",
                "site",
                "assignee",
                "priority",
                "priorityCode",
                "status",
                "statusCode",
                "dueDatetime",
                "audit",
                "auditId",
                "linkedToItem",
                "linkedToItemId",
                "creatorName",
                "creatorId",
                "createdDatetime",
                "modifiedDatetime",
                "completedDatetime",
            ]
        )
    for action in actions_array:
        actions_list = transform_action_object_to_list(action)
        actions_csv_wr.writerow(actions_list)
        del actions_list


def export_actions(logger, settings, sc_client, get_started):
    """
    Export all actions created after date specified
    :param logger:      The logger
    :param settings:    Settings from command line and configuration file
    :param sc_client:   instance of safetypy.SafetyCulture class

    """

    logger.info("Exporting iAuditor actions")
    last_successful_actions_export = get_last_successful_actions_export(
        logger, settings[CONFIG_NAME]
    )
    actions_array = sc_client.get_audit_actions(last_successful_actions_export)
    if actions_array is not None:
        logger.info("Found " + str(len(actions_array)) + " actions")
        if not get_started:
            save_exported_actions_to_csv_file(
                logger, settings[EXPORT_PATH], actions_array, settings[CONFIG_NAME]
            )
        else:
            save_exported_actions_to_db(logger, actions_array, settings, get_started)
        utc_iso_datetime_now = datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%S.000Z")
        update_actions_sync_marker_file(
            logger, utc_iso_datetime_now, settings[CONFIG_NAME]
        )
