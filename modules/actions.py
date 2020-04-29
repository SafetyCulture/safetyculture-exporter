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

def export_actions(logger, settings, sc_client, get_started):
    """
    Export all actions created after date specified
    :param logger:      The logger
    :param settings:    Settings from command line and configuration file
    :param sc_client:   instance of safetypy.SafetyCulture class
    """

    logger.info('Exporting iAuditor actions')
    last_successful_actions_export = get_last_successful_actions_export(logger)
    actions_array = sc_client.get_audit_actions(last_successful_actions_export)
    if actions_array is not None:
        logger.info('Found ' + str(len(actions_array)) + ' actions')
        if not get_started:
            save_exported_actions_to_csv_file(logger, settings[EXPORT_PATH], actions_array)
        else:
            save_exported_actions_to_db(logger, actions_array, settings, get_started)
        utc_iso_datetime_now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%S.000Z')
        update_actions_sync_marker_file(logger, utc_iso_datetime_now)


