import trainingpy as tp
import coloredlogs, logging
import errno
import os
from datetime import datetime
import sys
import pandas as pd
import safetypy as sp
import configparser

# Possible values here are DEBUG, INFO, WARN, ERROR and CRITICAL
LOG_LEVEL = logging.DEBUG


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
    log_filename = datetime.now().strftime('%Y-%m-%d') + '.log'
    exporter_logger = logging.getLogger('exporter_logger')
    exporter_logger.setLevel(LOG_LEVEL)
    formatter = logging.Formatter('%(asctime)s : %(levelname)s : %(message)s')

    fh = logging.FileHandler(filename=os.path.join(path_to_log_directory, log_filename))
    fh.setLevel(LOG_LEVEL)
    fh.setFormatter(formatter)
    exporter_logger.addHandler(fh)

    sh = logging.StreamHandler(sys.stdout)
    sh.setLevel(LOG_LEVEL)
    sh.setFormatter(formatter)
    exporter_logger.addHandler(sh)


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
            log_critical_error(logger, ex, 'An error happened trying to create ' + path)
            raise


def configure_logger():
    """
    Declare and validate existence of log directory; create and configure logger object

    :return:  instance of configured logger object
    """
    log_dir = os.path.join(os.getcwd(), 'log')
    create_directory_if_not_exists(None, log_dir)
    configure_logging(log_dir)
    logger = logging.getLogger('exporter_logger')
    coloredlogs.install(logger=logger)
    return logger


def users_to_dict(logger):
    # Returns a list of one user per row along with their team information.
    logger.info('Downloading User List')
    list_of_teams = tt.list_teams()
    logger.info('Found {} teams'.format(len(list_of_teams)))
    combined_list = []

    for team_id in list_of_teams:
        logger.info('Downloading users from {}'.format(team_id['title']))
        row = team_id
        row['TeamID'] = row.pop('id')
        row['TeamName'] = row.pop('title')
        row['TeamDescription'] = row.pop('description')
        id = row['TeamID']
        team_members = tt.list_teams_users(id)
        if team_members:
            for user in team_members:
                logger.info('Downloading {}'.format(user['email']))
                user['UserID'] = user.pop('id')
                iauditor_id = sc.get_id_from_email(user['email'])
                user['iAuditor UserID'] = iauditor_id
                user.update(row)
                combined_list.append(user)

    return combined_list


def lesson_records_to_dict(logger):
    # Returns the completed lessons in a dict
    user_list = users_to_dict(logger)
    combined_list = []
    for user in user_list:

        id = user['UserID']
        user_records = tt.discover_lesson_records(user_id=id)
        logger.info('Downloading {} Records for {}'.format(len(user_records),user['email']))
        for record in user_records:
            record['UserID'] = id
            combined_list.append(record)
    return combined_list


def main():
    logger = configure_logger()
    users = users_to_dict(logger)
    lesson_records = lesson_records_to_dict(logger)
    courses = tt.discover_courses()
    df_courses = pd.DataFrame(courses)
    df_users = pd.DataFrame(users)
    df_records = pd.DataFrame(lesson_records)
    print(df_users)
    print(df_courses)
    print(df_records)

    df_users.to_csv('users.csv')
    df_courses.to_csv('courses.csv')
    df_records.to_csv('records.csv')


config = configparser.ConfigParser()
config.read('ttconfig.ini')
tt_api_key = config['TOKENS']['TeamTrain']
tt = tp.Training(tt_api_key)
sc_api_key = config['TOKENS']['iAuditor']
sc = sp.SafetyCulture(sc_api_key)
main()

