# coding=utf-8
# Author: SafetyCulture
# Copyright: © SafetyCulture 2016
import sys
import time

from rich import print
from tqdm import tqdm

from iauditor_exporter.modules import csvExporter

try:
    from iauditor_exporter.modules.exporters import (
        export_audit_pdf_word,
        export_audit_json,
        export_audit_csv,
        export_actions,
    )
    from iauditor_exporter.modules.global_variables import *
    from iauditor_exporter.modules.last_successful import (
        get_last_successful,
        update_sync_marker_file,
    )
    from iauditor_exporter.modules.logger import configure_logger
    from iauditor_exporter.modules.media import (
        check_if_media_sync_offset_satisfied,
        export_audit_media,
    )
    from iauditor_exporter.modules.other import show_preferences_and_exit
    from iauditor_exporter.modules.settings import (
        parse_export_filename,
        parse_command_line_arguments,
        configure,
        create_directory_if_not_exists,
    )
    from iauditor_exporter.modules.sql import (
        sql_setup,
        end_session,
        query_max_last_modified,
        SQL_HEADER_ROW,
        bulk_import_sql,
        db_formatter,
    )
    from iauditor_exporter.modules.web_report_links import export_audit_web_report_link

except ImportError as e:
    print(e)
    print(
        "The ModuleNotFoundError indicates that some packages required by the script have not been installed. \n The "
        "error above will give details of whichever package was found to be missing first.\n Sometimes you need to "
        "close and reopen your command window after install, so try that first.\n "
        "If you continue to see this error, please review this page of the documentation: "
        "https://safetyculture.github.io/iauditor-exporter/script-setup/installing-packages/"
    )
    sys.exit()


def create_media_sync_log(logger, settings):
    path = os.path.join("media_sync_log")
    create_directory_if_not_exists(logger, path)
    contents_of_directory = os.listdir(path)
    filename = f"media_sync_ids-{settings[CONFIG_NAME]}.txt"
    file_name = os.path.join(path, filename)

    return filename, file_name, contents_of_directory


def create_file_if_not_exists(logger, file_name):
    logger.debug(f"Creating {file_name}")
    open(file_name, "a").close()


def inspections_to_process_next_sync(logger, audit, settings):
    filename, file_name, contents_of_directory = create_media_sync_log(logger, settings)
    if contents_of_directory:
        if filename in contents_of_directory:
            audit_ids_to_extend = open(file_name, "r")
            stripped_list = [line.strip() for line in audit_ids_to_extend.readlines()]
            audit_ids_to_extend.close()
            if audit["audit_id"] not in stripped_list:
                append_line_to_text_file(logger, file_name, audit["audit_id"])
            else:
                pass
    else:
        create_file_if_not_exists(logger, file_name)
        append_line_to_text_file(logger, file_name, audit["audit_id"])


def append_skipped_inspections(logger, list_of_audits, settings):
    """
    :param logger: Logging
    :param list_of_audits: The list of inspections from the API
    :param settings: Exporter Settings

    :return: API List with any media sync IDs appended
    """
    filename, file_name, contents_of_directory = create_media_sync_log(logger, settings)
    if contents_of_directory:
        if filename in contents_of_directory:
            audit_ids_to_extend = open(file_name, "r")
            list_to_extend = [line.strip() for line in audit_ids_to_extend.readlines()]
            list_of_audits["audits"] = [
                obj
                for obj in list_of_audits["audits"]
                if obj["audit_id"] not in list_to_extend
            ]
            list_of_audits["audits"].extend(
                {"audit_id": x, "modified_at": None} for x in list_to_extend
            )
            logger.info(
                "Found previously skipped inspections, will attempt to process again this run."
            )
    return list_of_audits


def is_file_empty(logger, file_name):
    """ Check if file is empty by reading first character in it"""
    with open(file_name, "r") as read_obj:
        one_char = read_obj.read(1)
        if not one_char:
            logger.debug("Media sync file is empty")
            return True
    logger.debug("Media sync file is populated")
    return False


def remove_id_from_media_sync_file(logger, settings, audit):
    filename, file_name, contents_of_directory = create_media_sync_log(logger, settings)
    if contents_of_directory:
        if filename in contents_of_directory:
            audit_ids_to_extend = open(file_name, "r")
            list_to_extend = [line.strip() for line in audit_ids_to_extend.readlines()]
            changed = False
            for i, audit_id in enumerate(list_to_extend):
                if audit["audit_id"] == audit_id:
                    changed = True
                    del list_to_extend[i]
            if changed:
                new_file = open(file_name, "w+")
                new_file.writelines(l + "\n" for l in list_to_extend)
                new_file.close()
                logger.debug(
                    f'Removed {audit["audit_id"]} from the media sync log as it was successfully exported.'
                )


def append_line_to_text_file(logger, text_file, line):
    logger.info(
        f"Writing {line.strip()} to media sync log so it can be exported later."
    )
    new_file = is_file_empty(logger, text_file)
    with open(text_file, "a") as file_object:
        if not new_file:
            file_object.write("\n")
        file_object.write(str(line).strip())
        file_object.close()


def sync_exports(logger, settings, sc_client):
    """
    Perform sync, exporting documents modified since last execution

    :param logger:    the logger
    :param settings:  Settings from command line and configuration file
    :param sc_client: Instance of SDK object
    """
    get_started = None
    if settings[EXPORT_ARCHIVED] is not None:
        archived_setting = settings[EXPORT_ARCHIVED]
    else:
        archived_setting = False
    if settings[EXPORT_COMPLETED] is not None:
        completed_setting = settings[EXPORT_COMPLETED]
    else:
        completed_setting = True
    if "actions-sql" in settings[EXPORT_FORMATS]:
        get_started = sql_setup(logger, settings, "actions")
        export_actions(logger, settings, sc_client, get_started)
    if "actions" in settings[EXPORT_FORMATS]:
        get_started = None
        get_started = None
        export_actions(logger, settings, sc_client, get_started)
    if not bool(
        set(settings[EXPORT_FORMATS])
        & {
            "pdf",
            "docx",
            "csv",
            "media",
            "web-report-link",
            "json",
            "sql",
            "pickle",
            "doc_creation",
        }
    ):
        return
    if "sql" in settings[EXPORT_FORMATS]:
        last_successful = get_last_successful(logger, settings[CONFIG_NAME])
    else:
        last_successful = get_last_successful(logger, settings[CONFIG_NAME])
    if settings[TEMPLATE_IDS] is not None:
        if settings[TEMPLATE_IDS].endswith(".txt"):
            file = settings[TEMPLATE_IDS].strip()
            f = open(file, "r")
            ids_to_search = []
            for id in f:
                ids_to_search.append(id.strip())
        elif len(settings[TEMPLATE_IDS]) != 1:
            ids_to_search = settings[TEMPLATE_IDS].split(",")
        else:
            ids_to_search = [settings[TEMPLATE_IDS][0]]
        list_of_audits = sc_client.discover_audits(
            modified_after=last_successful,
            template_id=ids_to_search,
            completed=completed_setting,
            archived=archived_setting,
        )
    else:
        list_of_audits = sc_client.discover_audits(
            modified_after=last_successful,
            completed=completed_setting,
            archived=archived_setting,
        )

    if list_of_audits is not None:
        if len(list_of_audits["audits"]) > 1000:
            dupe_removal = []
            new_audits = []
            show_dupes = []
            for audit in list_of_audits["audits"]:
                if audit["audit_id"] not in dupe_removal:
                    dupe_removal.append(audit["audit_id"])
                    new_audits.append(audit)
                else:
                    show_dupes.append(audit)
            list_of_audits["audits"] = new_audits

        list_of_audits = append_skipped_inspections(logger, list_of_audits, settings)

        logger.info(str(len(list_of_audits["audits"])) + " audits discovered")
        # logger.info(str(list_of_audits["total"]) + " audits discovered")
        export_count = 1
        export_total = len(list_of_audits["audits"])
        # export_total = list_of_audits["total"]
        get_started = "ignored"
        audits_to_process = list_of_audits["audits"]

        # chunks_to_process takes our long list of inspections and splits into smaller chunks. This helps
        # with the rate limiting, but more important means that if anything goes wrong, you don't always go
        # back to square one. From testing, 500 is a good number to use here and I do not recommend
        # going higher.

        per_chunk = settings["chunks"]

        if "sql" in settings[EXPORT_FORMATS]:
            get_started = sql_setup(logger, settings, "audit")

        elif "csv" in settings[EXPORT_FORMATS]:
            get_started = "csv"

        chunks_to_process = sc_client.chunks(audits_to_process, per_chunk)
        loop_through_chunks(
            chunks_to_process,
            logger,
            settings,
            sc_client,
            export_count,
            export_total,
            get_started,
            per_chunk,
        )

        if "sql" in settings[EXPORT_FORMATS]:
            end_session(get_started[1])
    else:
        logger.error(
            "There was a problem obtaining a list of inspections from the API. If the error above includes "
            '"unauthorized", you will need to generate a new API token - just re-run the script with the --setup '
            "parameter. If there is mention of a bad request, "
            "double check the formatting of your last successful file and config file. "
        )


def loop_through_chunks(
    chunks_to_process,
    logger,
    settings,
    sc_client,
    export_count,
    export_total,
    get_started,
    per_chunk,
):
    for chunk in chunks_to_process:
        if per_chunk > export_total:
            logger.info(f"Downloading {str(export_total)} total inspections...")
        else:
            logger.info(f"Downloading inspection data...")
        audits_to_process = sc_client.raise_pool(chunk)
        all_audits = []
        audit_pbar = tqdm(
            audits_to_process, total=export_total, initial=export_count - 1
        )
        debug_code = ""
        modified_at = ""
        for audit in audit_pbar:
            if audit:
                audit_pbar.set_description(f"Processing {audit['audit_id']}")
                logger.debug(
                    "Processing "
                    + str(audit["audit_id"])
                    + " - "
                    + str(export_count)
                    + "/"
                    + str(export_total)
                    + ")"
                )
                debug_code, modified_at = process_audit(
                    logger,
                    settings,
                    sc_client,
                    audit,
                    get_started,
                    all_audits=all_audits,
                )
                export_count += 1
        if "sql" in settings[EXPORT_FORMATS]:
            bulk_import_sql(logger, all_audits, get_started)
        if debug_code is not None and modified_at is not None:
            logger.debug(debug_code)
            update_sync_marker_file(modified_at, settings[CONFIG_NAME])


def process_audit(logger, settings, sc_client, audit, get_started, all_audits=[]):
    """
    Export audit in the format specified in settings. Formats include PDF, JSON, CSV, MS Word (docx), media, or
    web report link.
    :param get_started:
    :param all_audits:
    :param logger:      The logger
    :param settings:    Settings from command line and configuration file
    :param sc_client:   instance of safetypy.SafetyCulture class
    :param audit:       Audit JSON to be exported
    """

    if not check_if_media_sync_offset_satisfied(logger, settings, audit):
        debug_code = None
        modified_at = None
        inspections_to_process_next_sync(logger, audit, settings)
        return debug_code, modified_at

    audit_id = audit["audit_id"]
    audit_json = audit
    template_id = audit_json["template_id"]
    preference_id = None
    if (
        settings[PREFERENCES] is not None
        and template_id in settings[PREFERENCES].keys()
    ):
        preference_id = settings[PREFERENCES][template_id]
    export_filename = (
        parse_export_filename(audit_json, settings[FILENAME_ITEM_ID]) or audit_id
    )
    for export_format in settings[EXPORT_FORMATS]:
        if export_format == "sql":
            if get_started[0] == "complete":
                db_formatter(settings, audit, all_audits=all_audits)
            elif get_started[0] != "complete":
                logger.error(
                    "Something went wrong connecting to the database, please check your settings."
                )
                sys.exit(1)
        elif export_format == "csv":
            export_audit_csv(settings, audit_json)
        elif export_format in ["pdf", "docx"]:
            export_audit_pdf_word(
                logger,
                sc_client,
                settings,
                audit_id,
                preference_id,
                export_format,
                export_filename,
            )
        elif export_format == "json":
            export_audit_json(logger, settings, audit_json, export_filename)
        elif export_format == "media":
            export_audit_media(
                logger, sc_client, settings, audit_json, audit_id, export_filename
            )
        elif export_format == "web-report-link":
            export_audit_web_report_link(
                logger, settings, sc_client, audit_json, audit_id, template_id
            )
    debug_code = "setting last modified to " + audit["modified_at"]
    modified_at = audit["modified_at"]

    remove_id_from_media_sync_file(logger, settings, audit)

    return debug_code, modified_at


def loop(logger, sc_client, settings):
    """
    Loop sync until interrupted by user
    :param logger:     the logger
    :param sc_client:  instance of SafetyCulture SDK object
    :param settings:   dictionary containing config settings values
    """
    sync_delay_in_seconds = settings[SYNC_DELAY_IN_SECONDS]
    while True:
        sync_exports(logger, settings, sc_client)
        logger.info(
            "Next check will be in "
            + str(sync_delay_in_seconds)
            + " seconds. Waiting..."
        )
        time.sleep(sync_delay_in_seconds)


def main():
    try:
        logger = configure_logger()
        (
            path_to_config_file,
            export_formats,
            preferences_to_list,
            loop_enabled,
            docker_enabled,
            chunks,
        ) = parse_command_line_arguments(logger)
        sc_client, settings = configure(
            logger, path_to_config_file, export_formats, docker_enabled, chunks
        )
        if preferences_to_list is not None:
            show_preferences_and_exit(preferences_to_list, sc_client)
        if loop_enabled:
            loop(logger, sc_client, settings)
        else:
            sync_exports(logger, settings, sc_client)
            logger.info("Completed sync process, exiting")

    except KeyboardInterrupt:
        print("Interrupted by user, exiting.")
        sys.exit(0)


if __name__ == "__main__":
    main()
