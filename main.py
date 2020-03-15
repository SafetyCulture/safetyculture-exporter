import sys
from export_5S import internet_on, create_directories, export_audits_csv, update_database, export_actions_to_excel, \
    send_mail
from retrieve import retrieve_new_audits, retrieve_all_actions


def main():
    # User inputs
    token = '9fca4231efdec941bbb3bf1065ffbbc0965b8dbe9241ca435693d2d010dce877'
    templates_id = 'template_d3b1f799740443f28fe51a640fa19436,template_7f6e9a0e8832484eb3cc4f14129d82e1'

    internet = internet_on()

    # If the user has a valid internet connection
    if internet is False:
        print("\nCan't connect to internet, please check your connection\n")

    else:
        try:
            # Create exports and medias directories
            exports_folder_path, medias_folder_path = create_directories()
            # Retrieve audits
            audit_list_json = retrieve_new_audits(api_token=token, templates_id=templates_id, completed=True)
            # Export audits
            export_audits_csv(audit_list_json, exports_folder_path)
            # Updating excel database from exported files
            update_database(exports_folder_path, medias_folder_path, token)
            # Retrieve actions
            action_list = retrieve_all_actions(api_token=token)
            # Export actions
            export_actions_to_excel(action_list)
            # Send mail with Excel exports
            send_mail()

        # If user interupts process
        except KeyboardInterrupt:
            print("\nInterrupted by user, exiting")
            sys.exit(0)

        # If access denied by proxy
        except ConnectionError:
            print("\nAutorization to retrieve data from API denied, exiting")
            sys.exit(0)


if __name__ == '__main__':
    main()
