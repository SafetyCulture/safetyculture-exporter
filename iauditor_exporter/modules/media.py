import os
import shutil
from datetime import datetime, timedelta

import dateutil
import pytz

from iauditor_exporter.modules.global_variables import (
    MEDIA_SYNC_OFFSET_IN_SECONDS,
    EXPORT_PATH,
)
from iauditor_exporter.modules.logger import log_critical_error


def save_exported_media_to_file(logger, export_dir, media_file, filename, extension):
    """
    Write exported media item to disk at specified location with specified file name.
    Any existing file with the same name will be overwritten.
    :param logger:      the logger
    :param export_dir:  path to directory for exports
    :param media_file:  media file to write to disc
    :param filename:    filename to give exported image
    :param extension:   extension to give exported image
    """
    if not os.path.exists(export_dir):
        logger.info("Creating directory at {0} for media files.".format(export_dir))
        os.makedirs(export_dir)
    file_path = os.path.join(export_dir, filename + "." + extension)
    if os.path.isfile(file_path):
        logger.info("Overwriting existing report at " + file_path)
    try:
        with open(file_path, "wb") as out_file:
            shutil.copyfileobj(media_file.raw, out_file)
        del media_file
    except Exception as ex:
        log_critical_error(
            logger, ex, "Exception while writing" + file_path + " to file"
        )


def check_if_media_sync_offset_satisfied(logger, settings, audit):
    """
    Check if the media sync offset is satisfied. The media sync offset is a duration in seconds specified in the
    configuration file. This duration is the amount of time audit media is given to sync up with SafetyCulture servers
    before this tool exports the audit data.
    :param logger:    The logger
    :param settings:  Settings from command line and configuration file
    :param audit:     Audit JSON
    :return:          Boolean - True if the media sync offset is satisfied, otherwise, returns false.
    """
    modified_at = dateutil.parser.parse(audit["modified_at"])
    now = datetime.utcnow()
    elapsed_time_difference = pytz.utc.localize(now) - modified_at
    # if the media_sync_offset has been satisfied
    if not elapsed_time_difference > timedelta(
        seconds=settings[MEDIA_SYNC_OFFSET_IN_SECONDS]
    ):
        logger.info(
            "Audit {0} modified too recently, some media may not have completed syncing. "
            "Skipping export until next sync cycle".format(audit["audit_id"])
        )
        return False
    return True


def export_audit_media(
    logger, sc_client, settings, audit_json, audit_id, export_filename
):
    """
    Save audit media files to disk
    :param logger:      The logger
    :param sc_client:   instance of safetypy.SafetyCulture class
    :param settings:    Settings from command line and configuration file
    :param audit_json:  Audit JSON
    :param audit_id:    Unique audit UUID
    :param export_filename:     String indicating what to name the exported audit file
    """
    media_export_path = os.path.join(settings[EXPORT_PATH], "media", export_filename)
    media_id_list = get_media_from_audit(logger, audit_json)
    for media_id in media_id_list:
        extension = media_id[1]
        media_id = media_id[0]
        if not extension:
            extension = "jpg"
        media_file = sc_client.get_media(audit_id, media_id)
        if media_file is None:
            logger.warn("Failed to save media object {0}".format(media_id))
            continue
        logger.info("Saving media_{0} to disc.".format(media_id))
        media_export_filename = media_id
        save_exported_media_to_file(
            logger, media_export_path, media_file, media_export_filename, extension
        )


def get_media_from_audit(logger, audit_json):
    """
    Retrieve media IDs from a audit JSON
    :param logger: the logger
    :param audit_json: single audit JSON
    :return: list of media IDs
    """
    media_id_list = []
    for item in audit_json["header_items"] + audit_json["items"]:
        # This condition checks for media attached to question and media type fields.
        if "media" in item.keys():
            for media in item["media"]:
                if "file_ext" in media.keys():
                    file_ext = media["file_ext"]
                else:
                    file_ext = "jpg"
                media_id_list.append([media["media_id"], file_ext])
        # This condition checks for media attached to signature and drawing type fields.
        if "responses" in item.keys() and "image" in item["responses"].keys():
            if "file_ext" in item["responses"]["image"].keys():
                file_ext = item["responses"]["image"]["file_ext"]
            else:
                file_ext = "jpg"
            media_id_list.append([item["responses"]["image"]["media_id"], file_ext])
        # This condition checks for media attached to information type fields.
        if "options" in item.keys() and "media" in item["options"].keys():
            if "file_ext" in item["options"]["media"].keys():
                file_ext = item["options"]["media"]["file_ext"]
            else:
                file_ext = "jpg"
            media_id_list.append([item["options"]["media"]["media_id"], file_ext])
    logger.info(
        "Discovered {0} media files associated with {1}.".format(
            len(media_id_list), audit_json["audit_id"]
        )
    )
    return media_id_list
