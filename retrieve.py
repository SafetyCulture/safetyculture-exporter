import testpy as sp
import os
from datetime import datetime
from dateutil.relativedelta import relativedelta


def retrieve_all_audits(api_token, templates_id, modified_after='2019-01-01', completed=False):
    """
    Retrieve all available audits on user iAuditor profile from chosen templates in JSON format
    :param api_token:         user API token (str)
    :param templates_id:      list of template id (list)
    :param modified_after:    oldest modified audit date considered when retrieving audit (str)
    :param completed:         retrieve only completed audits (bool)
    :return audit_list_json:  list of audits in json format (list)
    """
    sc_client = api_connection(api_token)
    audit_list = discover_audits(sc_client, templates_id, modified_after, completed)
    audit_list_json = get_audits_json(sc_client, audit_list)
    return audit_list_json


def retrieve_all_actions(api_token, modified_after='2019-01-01'):
    """
    Retrieve all available audits on user iAuditor profile from chosen templates in JSON format
    :param api_token:         user API token (str)
    :param modified_after:    oldest modified audit date considered when retrieving action (str)
    :return action_list:      list of actions (list)
    """
    sc_client = api_connection(api_token)
    action_list = sc_client.get_audit_actions(modified_after)
    return action_list


def retrieve_list_audits(api_token, audit_list):
    """
    Retrieve audits from list of audits in JSON format
    :param api_token:         user API token (str)
    :param audit_list:        list of audit id (list)
    :return audit_list_json:  list of audits in json format (list)
    """
    sc_client = api_connection(api_token)
    audit_list_json = get_audits_json(sc_client, audit_list)
    return audit_list_json


def retrieve_new_audits(api_token, templates_id, completed=False, path=str(os.getcwd()) + '/exports', week_number=1):
    """
    Retrieve new available audits on user iAuditor profile from chosen templates in JSON format or audits which have
    been modified in the last chosen number of week(s)
    :param week_number:
    :param api_token:          user API token (str)
    :param templates_id:       list of template id (list)
    :param completed:          retrieve only completed audits (bool)
    :param path:               path to check if audit already exists (str)
    :param week_number:        number of considered week(s) from today (int)
    :return audit_list_json:   list of audits in json format (list)
    """
    sc_client = api_connection(api_token)
    audit_list = discover_audits(sc_client, templates_id, modified_after='2019-01-01', completed=completed)
    new_audit_list = []
    for template in audit_list:
        for audit in template['audits']:
            if str(audit['audit_id']) + '.csv' not in os.listdir(path) or \
                    datetime.strptime(str(audit['modified_at'][:10]), '%Y-%m-%d') \
                    + relativedelta(weeks=week_number) >= datetime.now():
                new_audit_list.append(str(audit['audit_id']))
    audit_list_json = get_audits_json(sc_client, new_audit_list)
    return audit_list_json


def api_connection(api_token):
    """
    Establish connection to iAuditor API
    :param api_token:    user API token (str)
    :return sc_client:   instance of SafetyCulture SDK object (obj)
    """
    sc_client = sp.SafetyCulture(api_token, proxy_settings=None, certificate_settings=None)
    return sc_client


def discover_audits(sc_client, template_id, modified_after, completed):
    """
    Get list of audits
    :param sc_client:        instance of SafetyCulture SDK object (obj)
    :param template_id:      list of template id (list)
    :param modified_after:   oldest modified audit date considered when retrieving audit (str)
    :param completed:        retrieve only completed audits (bool)
    :return audit_list:      list of audits in list of template (list)
    """
    audit_list = []
    for i in range(len(template_id)):
        audit_list.append(sc_client.discover_audits(template_id[i], modified_after, completed))
    return audit_list


def get_audits_json(sc_client, audit_list):
    """
    From list of audits get list of json audits
    :param sc_client:         instance of SafetyCulture SDK object (obj)
    :param audit_list:        list of audits (list)
    :return audit_list_json:  list of audits in json format (list)
    """
    audit_number = audit_count(audit_list)
    print('\nRetrieving {} audits\n'.format(audit_number))
    audit_list_json = []
    try:
        for template in audit_list:
            for audit in template['audits']:
                audit_list_json.append(sc_client.get_audit(audit['audit_id']))
        return audit_list_json
    except TypeError:
        try:
            for audit in audit_list:
                audit_list_json.append(sc_client.get_audit(audit['audit_id']))
            return audit_list_json
        except TypeError:
            for audit in audit_list:
                audit_list_json.append(sc_client.get_audit(audit))
            return audit_list_json


def audit_count(audit_list):
    """
    Get the number of audits to retrieve
    :param audit_list:      list of audits (list)
    :return audit_number:   number of audits (int)
    """
    try:
        audit_number = 0
        for template in audit_list:
            audit_number = audit_number + len(template['audits'])
        return audit_number
    except TypeError:
        audit_number = len(audit_list)
        return audit_number
