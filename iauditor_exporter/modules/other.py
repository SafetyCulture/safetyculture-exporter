import sys


def show_preferences_and_exit(list_preferences, sc_client):
    """
    Display preferences to stdout and exit

    :param list_preferences: empty list for all preference, list of template_ids if specified at command line
    :param sc_client:            instance of SDK object, used to retrieve preferences
    """
    row_boundary = "|" + "-" * 136 + "|"
    row_format = "|{0:<37} | {1:<40} | {2:<10}| {3:<10}|"
    print(row_boundary)
    print(row_format.format("Preference ID", "Preference Name", "Global", "Default"))
    print(row_boundary)

    if len(list_preferences) > 0:
        for template_id in list_preferences:
            preferences = sc_client.get_preference_ids(template_id)
            for preference in preferences["preferences"]:
                preference_id = str(preference["id"])
                preference_name = str(preference["label"])[:35]
                is_global = str(preference["is_global"])
                is_default = str(preference["is_default"])
                print(
                    row_format.format(
                        preference_id, preference_name, is_global, is_default
                    )
                )
                print(row_boundary)
        sys.exit()
    else:
        preferences = sc_client.get_preference_ids()
        for preference in preferences["preferences"]:
            preference_id = str(preference["id"])
            preference_name = str(preference["label"])[:35]
            is_global = str(preference["is_global"])
            is_default = str(preference["is_default"])
            print(
                row_format.format(preference_id, preference_name, is_global, is_default)
            )
            print(row_boundary)
        sys.exit(0)
