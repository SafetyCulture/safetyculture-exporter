# import os
# import unicodecsv as csv
import csvExporter

# from datetime import datetime
# from modules.global_variables import EXPORT_PATH, ACTIONS_EXPORT_FILENAME, CONFIG_NAME
# from modules.last_successful import get_last_successful_actions_export, update_actions_sync_marker_file
# from modules.sql import save_exported_actions_to_db


def transform_action_object_to_list(action):
    priority_codes = {0: 'None', 10: 'Low', 20: 'Medium', 30: 'High'}
    status_codes = {0: 'To Do', 10: 'In Progress', 50: 'Done', 60: 'Cannot Do'}
    get_json_property = csvExporter.get_json_property
    actions_list = [get_json_property(action, 'action_id'), get_json_property(action, 'description')]
    assignee_list = []
    for assignee in get_json_property(action, 'assignees'):
        assignee_list.append(get_json_property(assignee, 'name'))
    actions_list.append(", ".join(assignee_list))
    actions_list.append(get_json_property(priority_codes, get_json_property(action, 'priority')))
    actions_list.append(get_json_property(action, 'priority'))
    actions_list.append(get_json_property(status_codes, get_json_property(action, 'status')))
    actions_list.append(get_json_property(action, 'status'))
    actions_list.append(get_json_property(action, 'due_at'))
    actions_list.append(get_json_property(action, 'audit', 'name'))
    actions_list.append(get_json_property(action, 'audit', 'audit_id'))
    actions_list.append(get_json_property(action, 'item', 'label'))
    actions_list.append(get_json_property(action, 'item', 'item_id'))
    actions_list.append(get_json_property(action, 'created_by', 'name'))
    actions_list.append(get_json_property(action, 'created_by', 'user_id'))
    actions_list.append(get_json_property(action, 'created_at'))
    actions_list.append(get_json_property(action, 'modified_at'))
    actions_list.append(get_json_property(action, 'completed_at'))
    return actions_list


