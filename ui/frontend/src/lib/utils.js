export function trim(value, size) {
    if (value.length > size) {
        return value.substring(0, size).concat(" ...")
    }
    return value
}

export function isNullOrEmptyObject(obj) {
    if(obj === null) {
        return true
    }
    return Object.keys(obj).length === 0;
}

export const allTables = ['inspections', 'inspection_items', 'schedules', 'templates', 'template_permissions',
    'sites', 'site_members', 'groups', 'group_users', 'schedule_assignees', 'schedule_occurrences', 'actions',
    'action_assignees', 'action_timeline_items', 'issues', 'issue_timeline_items', 'assets', 'users', 'issue_assignees'];
