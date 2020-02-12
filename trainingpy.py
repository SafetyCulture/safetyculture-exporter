import errno
from time import sleep

import requests
from getpass import getpass
import coloredlogs, logging
import json
import os
import re
import sys
from datetime import datetime

HTTP_USER_AGENT_ID = 'safetyculture-python-sdk'


def get_user_api_token(logger):
    """
    Generate iAuditor API Token
    :param logger:  the logger
    :return:        API Token if authenticated else None
    """
    username = input("TeamTrain email: ")
    password = getpass()
    generate_token_url = "https://sandpit-api.safetyculture.com/training/api/auth/token"
    post_body = {
        'email': username,
        'password': password
    }
    payload = json.dumps(post_body)
    headers = {
        'content-type': "application/json",
        'cache-control': "no-cache",
    }
    response = requests.request("POST", generate_token_url, data=payload, headers=headers)
    if response.status_code == requests.codes.ok:
        return response.json()['access_token']
    else:
        logger.error('An error occurred calling ' + generate_token_url + ': ' + str(response.json()))
        return None

class Training:
    def __init__(self, api_token):
        self.current_dir = os.getcwd()
        self.log_dir = self.current_dir + '/log/'
        self.api_url = 'https://sandpit-api.safetyculture.com/training/api/'
        self.courses_url = self.api_url + 'courses'
        self.lessons_url = self.api_url + 'records/lessons'
        self.teams_url = self.api_url + 'teams'
        self.users_url = self.api_url + 'users'

        self.create_directory_if_not_exists(self.log_dir)
        self.configure_logging()
        logger = logging.getLogger('tt_logger')
        try:
            if api_token:
                self.api_token = api_token
            else:
                logger.error('API token missing')
                self.api_token = None
        except Exception as ex:
            self.log_critical_error(ex, 'API token is missing or invalid. Exiting.')
            exit()
        if self.api_token:
            self.custom_http_headers = {
                'User-Agent': HTTP_USER_AGENT_ID,
                'Authorization': 'Bearer ' + self.api_token
            }
        else:
            logger.error('No valid API token parsed! Exiting.')
            sys.exit(1)

    def authenticated_request_get(self, url, params=None):
        sleep(2)
        return requests.get(url, headers=self.custom_http_headers, params=params)

    def authenticated_request_post(self, url, data):
        sleep(2)
        self.custom_http_headers['content-type'] = 'application/json'
        response = requests.post(url, data, headers=self.custom_http_headers)
        del self.custom_http_headers['content-type']
        return response

    def authenticated_request_put(self, url, data):
        sleep(2)
        self.custom_http_headers['content-type'] = 'application/json'
        response = requests.put(url, data, headers=self.custom_http_headers)
        del self.custom_http_headers['content-type']
        return response

    def authenticated_request_delete(self, url):
        sleep(2)
        return requests.delete(url, headers=self.custom_http_headers)

    @staticmethod
    def log_critical_error(ex, message):
        """
        Write exception and description message to log

        :param ex:       Exception instance to log
        :param message:  Descriptive message to describe exception
        """
        logger = logging.getLogger('tt_logger')

        if logger is not None:
            logger.critical(message)
            logger.critical(ex)

    @staticmethod
    def log_http_status(status_code, message):
        """
        Write http status code and descriptive message to log

        :param status_code:  http status code to log
        :param message:      to describe where the status code was obtained
        """
        logger = logging.getLogger('tt_logger')
        status_description = requests.status_codes._codes[status_code][0]
        log_string = str(status_code) + ' [' + status_description + '] status received ' + message
        logger.info(log_string) if status_code == requests.codes.ok else logger.error(log_string)

    def configure_logging(self):
        """
        Configure logging to log to std output as well as to log file
        """
        log_level = logging.WARNING

        log_filename = datetime.now().strftime('%Y-%m-%d') + '.log'
        sp_logger = logging.getLogger('tt_logger')
        sp_logger.setLevel(log_level)
        formatter = logging.Formatter('%(asctime)s : %(levelname)s : %(message)s')

        fh = logging.FileHandler(filename=self.log_dir + log_filename)
        fh.setLevel(log_level)
        fh.setFormatter(formatter)
        sp_logger.addHandler(fh)

        sh = logging.StreamHandler(sys.stdout)
        sh.setLevel(log_level)
        sh.setFormatter(formatter)
        sp_logger.addHandler(sh)

    def create_directory_if_not_exists(self, path):
        """
        Creates 'path' if it does not exist

        If creation fails, an exception will be thrown

        :param path:    the path to ensure it exists
        """
        try:
            os.makedirs(path)
        except OSError as ex:
            if ex.errno == errno.EEXIST and os.path.isdir(path):
                pass
            else:
                self.log_critical_error(ex, 'An error happened trying to create ' + path)
                raise

    def get_offset_items(self, search_url, params=None, offset=0, list_of_items=None):
        if list_of_items is None:
            list_of_items = []
        if params is None:
            params = {}
        params['offset'] = offset
        response = self.authenticated_request_get(search_url, params)
        result = response.json() if response.status_code == requests.codes.ok else None
        results_returned = len(result['items'])
        if results_returned == 20:
            offset += 20
            for course in result['items']:
                list_of_items.append(course)
            return self.get_offset_items(search_url, offset=offset, list_of_items=list_of_items)
        else:
            for course in result['items']:
                list_of_items.append(course)

        # self.log_http_status(response.status_code, log_message)

        return list_of_items

    def discover_courses(self):
        """
        Args:
            Returns a list of courses
            offset:
        """
        logger = logging.getLogger('tt_logger')

        search_url = self.courses_url
        list_of_courses = self.get_offset_items(search_url)

        return list_of_courses

    def get_course(self, course_id):
        if course_id:
            search_url = self.courses_url + '/' + course_id + '/lessons'
        else:
            return None
        response = self.authenticated_request_get(search_url)
        return response.json()

    def get_course_teams(self, course_id):
        if course_id:
            search_url = self.courses_url + '/' + course_id + '/teams'
        else:
            return None

        list_of_teams = self.get_offset_items(search_url)

        return list_of_teams

    def discover_lesson_records(self, course_id=None, lesson_id=None, user_id=None):
        search_url = self.lessons_url
        params = {}
        if course_id:
            params['course_id'] = course_id
        if lesson_id:
            params['lesson_id'] = lesson_id
        if user_id:
            params['user_id'] = user_id

        list_of_lesson_records = self.get_offset_items(search_url, params)

        return list_of_lesson_records

    def list_teams(self):
        search_url = self.teams_url
        list_of_teams = self.get_offset_items(search_url)

        return list_of_teams

    def list_teams_users(self, team_id):
        if team_id:
            search_url = self.teams_url + '/' + team_id + '/users?'
            list_of_team_users = self.get_offset_items(search_url)

            return list_of_team_users
        else:
            return None

    def list_users(self):
        search_url = self.users_url
        list_of_users = self.get_offset_items(search_url)

        return list_of_users
