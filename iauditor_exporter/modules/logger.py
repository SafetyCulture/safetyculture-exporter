import errno
import logging
import os
import sys
from datetime import datetime
import coloredlogs
from iauditor_exporter.modules.global_variables import LOG_LEVEL


def create_directory_if_not_exists(logger, path):
    """
    Creates 'path' if it does not exist

    If creation fails, an exception will be thrown

    :param logger:  the logger
    :param path:    the path to ensure it exists
    """
    try:
        os.makedirs(path)
    except OSError as ex:
        if ex.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            log_critical_error(logger, ex, "An error happened trying to create " + path)
            raise


def log_critical_error(logger, ex, message):
    """
    Logs the exception at 'CRITICAL' log level

    :param logger:  the logger
    :param ex:      exception to log
    :param message: descriptive message to log details of where/why ex occurred
    """
    if logger is not None:
        logger.critical(message)
        logger.critical(ex)


def configure_logging(path_to_log_directory):
    """
    Configure logger

    :param path_to_log_directory:  path to directory to write log file in
    :return:
    """
    coloredlogs.install(level="INFO")
    log_filename = datetime.now().strftime("%Y-%m-%d") + ".log"
    exporter_logger = logging.getLogger(__name__)
    exporter_logger.setLevel(LOG_LEVEL)
    formatter = logging.Formatter("%(asctime)s : %(levelname)s : %(message)s")

    fh = logging.FileHandler(filename=os.path.join(path_to_log_directory, log_filename))
    fh.setLevel(LOG_LEVEL)
    fh.setFormatter(formatter)
    exporter_logger.addHandler(fh)

    sh = logging.StreamHandler(sys.stdout)
    sh.setLevel(logging.FATAL)
    sh.setFormatter(formatter)
    exporter_logger.addHandler(sh)

    return exporter_logger


def configure_logger():
    """
    Declare and validate existence of log directory; create and configure logger object

    :return:  instance of configured logger object
    """
    log_dir = os.path.join(os.getcwd(), "log")
    create_directory_if_not_exists(None, log_dir)
    logger = configure_logging(log_dir)

    # logger = logging.getLogger(__name__)

    return logger
