import os

import unicodecsv as csv

import iauditor_exporter.modules.csvExporter as csvExporter
from iauditor_exporter.modules.global_variables import EXPORT_PATH
from iauditor_exporter.modules.logger import log_critical_error


def save_web_report_link_to_file(logger, export_dir, web_report_data):
    """
    Write Web Report links to 'web-report-links.csv' on disk at specified location
    Any existing file with the same name will be appended to
    :param logger:          the logger
    :param export_dir:      path to directory for exports
    :param web_report_data:     Data to write to CSV: Template ID, Template name, Audit ID, Audit name, Web Report link
    """
    if not os.path.exists(export_dir):
        logger.info(
            "Creating directory at {0} for Web Report links.".format(export_dir)
        )
        os.makedirs(export_dir)
    file_path = os.path.join(export_dir, "web-report-links.csv")
    if os.path.isfile(file_path):
        logger.info("Appending Web Report link to " + file_path)
        try:
            with open(file_path, "ab") as web_report_link_csv:
                wr = csv.writer(
                    web_report_link_csv, dialect="excel", quoting=csv.QUOTE_ALL
                )
                wr.writerow(web_report_data)
                web_report_link_csv.close()
        except Exception as ex:
            log_critical_error(
                logger, ex, "Exception while writing" + file_path + " to file"
            )
    else:
        logger.info("Creating " + file_path)
        logger.info("Appending web report to " + file_path)
        try:
            with open(file_path, "wb") as web_report_link_csv:
                wr = csv.writer(
                    web_report_link_csv, dialect="excel", quoting=csv.QUOTE_ALL
                )
                wr.writerow(
                    [
                        "Template ID",
                        "Template Name",
                        "Audit ID",
                        "Audit Name",
                        "Web Report Link",
                    ]
                )
                wr.writerow(web_report_data)
                web_report_link_csv.close()
        except Exception as ex:
            log_critical_error(
                logger, ex, "Exception while writing" + file_path + " to file"
            )


def export_audit_web_report_link(
    logger, settings, sc_client, audit_json, audit_id, template_id
):
    """
    Save web report link to disk in a CSV file.
    :param logger:      The logger
    :param sc_client:   instance of safetypy.SafetyCulture class
    :param settings:    Settings from command line and configuration file
    :param audit_json:  Audit JSON
    :param audit_id:    Unique audit UUID
    :param template_id: Unique template UUID
    """
    web_report_link = sc_client.get_web_report(audit_id)
    web_report_data = [
        template_id,
        csvExporter.get_json_property(audit_json, "template_data", "metadata", "name"),
        audit_id,
        csvExporter.get_json_property(audit_json, "audit_data", "name"),
        web_report_link,
    ]
    save_web_report_link_to_file(logger, settings[EXPORT_PATH], web_report_data)
