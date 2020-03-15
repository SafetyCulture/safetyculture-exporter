import errno
import os
import pandas as pd
from retrieve import api_connection
import csvExporter
import unicodecsv as csv
import math
import shutil
import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.mime.base import MIMEBase
from email import encoders
import datetime
try:
    import win32com.client as win32
except:
    pass
try:
    import httplib
except:
    import http.client as httplib


def update_database(exports_folder_path, medias_folder_path, api_token):
    """
    Update Excel database with exported csv files
    :param exports_folder_path:    path to exports folder (str)
    :param exports_folder_path:    path to medias folder (str)
    :param api_token:              user API token (str)
    :return df:                    dataframe
    """
    # Connecting to iAuditor API
    sc_client = api_connection(api_token)
    df = create_dataframe()
    audit_list = os.listdir(exports_folder_path)
    total_audit = len(audit_list)
    cpt = 1
    print('')
    for filename in audit_list:
        if not filename.startswith('.'):
            print('Exporting audit {}/{} to Excel database'.format(cpt, total_audit))
            df_new = audit_csv_to_dataframe(filename, exports_folder_path + '/', medias_folder_path, sc_client)
            df = concat_dataframes(df, df_new)
            cpt = cpt + 1

    df = end_dataframe(df)
    file = 'iAuditor_Exports.xlsx'
    df.to_excel(file, engine='xlsxwriter', sheet_name='Audits')

    print('\nAudits updated in iAuditor Excel database !\n')



def audit_csv_to_dataframe(File, exports_folder_path, medias_folder_path, sc_client):
    # Read File with pandas
    df_file = pd.read_csv(exports_folder_path + '/' + File)

    # Create one row datframe df
    col_names = ['audit_id', 'audit_author', 'audit_template_name', 'audit_site'
        , 'audit_sector', 'audit_area', 'audit_date', 'audit_score', 'audit_forecast', 'audit_month', 'audit_year',
                 'im_bonnes_pratiques_id', 'description_bonnes_pratiques', 'first_non_compliant_question']
    index = [1]
    df = pd.DataFrame(columns=col_names, index=index)

    # Retrieve values from the file dataframe
    try:
        df.loc[1, ['audit_id']] = df_file.loc[0, ['AuditID'][0]]
        df.loc[1, ['audit_author']] = df_file.loc[0, ['PreparedBy'][0]]
        df.loc[1, ['audit_template_name']] = df_file.loc[0, ['TemplateName'][0]]

        if (str(df_file.loc[1, ['Response'][0]])) == '' or str(df_file.loc[1, ['Response'][0]]) == 'nan':
            df.loc[1, ['audit_site']] = df_file.loc[1, ['AuditSite'][0]]
        else:
            df.loc[1, ['audit_site']] = df_file.loc[1, ['Response'][0]]

        date = df_file.loc[0, ['ConductedOn'][0]]
        df.loc[1, ['audit_date']] = pd.to_datetime(date.split("T", 1)[0], infer_datetime_format=True)
        audit_max_score = 5
        df.loc[1, ['audit_score']] = audit_max_score * float(
            df_file.loc[0, ['AuditScorePercentage'][0]]) / 100
    except:
        pass

    # Delete "SLS " in front of site name
    df['audit_site'] = df['audit_site'].str.replace(r'SLS - ', '')
    df['audit_site'] = df['audit_site'].str.replace(r'SLS ', '')
    df['audit_site'].replace({'Sendayan - Wheels and Brakes': 'Sendayan', 'Velizy - Engineering': 'Vélizy Engineering'
                                 , 'Velizy - Program': 'Vélizy Program', 'Velizy - Test lab': 'Vélizy Test Lab'
                                 , 'Gloucester - LGI': 'Gloucester', 'Velizy Engineering': 'Vélizy Engineering'
                                 , 'Walton – Carbon': 'Walton - Carbon', 'Bidos - LGI': 'Bidos'}, inplace=True)

    # Retrieve Area value from dataframe
    string1 = '{} - Area'.format(df.loc[1, ['audit_site']][0])
    string2 = '{} Area'.format(df.loc[1, ['audit_site']][0])
    try:
        index = df_file['Label'].loc[lambda x: x == string1].index[0]
        area = df_file.loc[index, ['Response']][0]
    except:
        try:
            index = df_file['Label'].loc[lambda x: x == string2].index[0]
            area = df_file.loc[index, ['Response']][0]
        except:
            area = ''
    try:
        df.loc[1, ['audit_area']] = area.split(" | ", 1)[1]
    except:
        try:
            df.loc[1, ['audit_area']] = area.split(" - ", 1)[0]
        except:
            df.loc[1, ['audit_area']] = area

    # Retrieve Sector value from dataframe
    string1 = '{} - Sector'.format(df.loc[1, ['audit_site']][0])
    string2 = '{} Sector'.format(df.loc[1, ['audit_site']][0])
    try:
        index = df_file['Label'].loc[lambda x: x == string1].index[0]
        sector = df_file.loc[index, ['Response']][0]
    except:
        try:
            index = df_file['Label'].loc[lambda x: x == string2].index[0]
            sector = df_file.loc[index, ['Response']][0]
        except:
            sector = ''
    try:
        df.loc[1, ['audit_sector']] = sector.split(" | ", 1)[1]
    except:
        try:
            df.loc[1, ['audit_sector']] = sector.split(" - ", 1)[0]
        except:
            df.loc[1, ['audit_sector']] = sector

    # Change sector names in dataframe
    df['audit_sector'].replace({
        'Procédés Spéciaux | Special Treatments': 'Special Treatments',
        'Usinage | Machining': 'Machining',
        'Communal Area | Communs': 'Communal Areas',
        'Communs': 'Communal Areas',
        'Machining | Usinage': 'Machining',
        'Usinage': 'Machining',
        'Special Treatments | Procédés Spéciaux': 'Special Treatments',
        'Procédés Spéciaux': 'Special Treatments',
        'Assembly | Montage': 'Assembly',
        'Montage | Assembly': 'Assembly',
        'Montage': 'Assembly',
        'Offices | Bureaux': 'Offices',
        'Bureaux': 'Offices',
        'Communal Area': 'Communal Areas',
        'Machining ': 'Machining'}, inplace=True)

    # Change site names in dataframe
    df['audit_site'].replace({
        'Nexon – Systems Equipment': 'Nexon - Systems Equipment'
        , 'Molsheim – Systems Equipment': 'Molsheim - Systems Equipment'}, inplace=True)

    # Retrieve First Non Compliant Question
    try:
        index = df_file['Response'].loc[lambda x: x == 'Non conforme'].index[0]
        df.loc[1, ['first_non_compliant_question']] = df_file.loc[index, ['Label'][0]]
    except:
        try:
            index = df_file['Response'].loc[lambda x: x == 'Non-Compliant'].index[0]
            df.loc[1, ['first_non_compliant_question']] = df_file.loc[index, ['Label'][0]]
        except:
            pass

    # Retrieve Forecast value from dataframe
    try:
        index = df_file['Label'].loc[lambda x: x == 'Quelle prévision pour le prochain audit ?'].index[0]
        index1 = df_file['Label'].loc[lambda x: x == 'What is your next audit score forecast?'].index[0]
        df.loc[1, ['audit_forecast']] = float(df_file.loc[index, ['Response']]) + float(
            df_file.loc[index1, ['Response']])
    except:
        try :
            answer = "EN - What is your next audit score forecast?\nFR - Quelle prévision (Note 5S) pour le prochain audit ?"
            index = df_file['Label'].loc[lambda x: x == answer].index[0]
            df.loc[1, ['audit_forecast']] = float(df_file.loc[index, ['Response']])
        except:
            df.loc[1, ['audit_forecast']] = 0

    # Add date infos
    try:
        df['audit_date'] = pd.to_datetime(df['audit_date'])
        df.loc[1, ['audit_month']] = df.loc[1, ['audit_date']].dt.month[0]
        df.loc[1, ['audit_year']] = df.loc[1, ['audit_date']].dt.year[0]
    except:
        pass

    # Retieve good pactrices image id
    try:
        # (french question)
        index = df_file['Label'].loc[
            lambda x: x == 'Choisissez la zone 5S à laquelle appartient la bonne pratique que vous partagez'].index
        for i in range(0, index.shape[0]):
            if df_file.loc[index[i], ['MediaHypertextReference']][0] != 'nan':
                media_id = str(df_file.loc[index[i], ['MediaHypertextReference']][0])
                df.loc[1, ['im_bonnes_pratiques_id']] = media_id.split('/', -1)[-1]
    except:
        try:
            # (english question)
            index = df_file['Label'].loc[lambda x: x == 'Best Practice'].index
            for i in range(0, index.shape[0]):
                if df_file.loc[index[i], ['MediaHypertextReference']][0] != 'nan':
                    media_id = str(df_file.loc[index[i], ['MediaHypertextReference']][0])
                    df.loc[1, ['im_bonnes_pratiques_id']] = media_id.split('/', -1)[-1]
        except:
            df.loc[1, ['im_bonnes_pratiques_id']] = ''
    df['im_bonnes_pratiques_id'].replace({'nan': ''}, inplace=True)

    # Retrieve good practices description
    try:
        # (french question)
        index = df_file['Label'].loc[
            lambda x: x == 'Décrivez la bonne pratique'].index
        for i in range(0, index.shape[0]):
            if df_file.loc[index[i], ['Response']][0] != 'nan':
                description = str(df_file.loc[index[i], ['Response']][0])
                df.loc[1, ['description_bonnes_pratiques']] = description
    except:
        try:
            # (english question)
            index = df_file['Label'].loc[lambda x: x == 'Best Practice'].index
            for i in range(0, index.shape[0]):
                if df_file.loc[index[i], ['Response']][0] != 'nan':
                    description = str(df_file.loc[index[i], ['Response']][0])
                    df.loc[1, ['description_bonnes_pratiques']] = description
        except:
            df.loc[1, ['description_bonnes_pratiques']] = ''
    df['description_bonnes_pratiques'].replace({'nan': ''}, inplace=True)

    # Retieve and export good practice media files
    try:
        if str(df['im_bonnes_pratiques_id'][1]) is not '' and math.isnan(float(df['im_bonnes_pratiques_id'][1])) is 1:
            pass
    except:
        df.loc[1, ['im_bonnes_pratiques_id']] = os.path.join('https://inshare.collab.group.safran/com/sls-acrfm/'
                                                        '5S%20News/5S%20Databases/medias/', media_id.split('/', -1)[-1]) + '.jpg'
        audit_id = df_file.loc[0, ['AuditID'][0]]
        media_file = sc_client.get_media(audit_id, media_id.split('/', -1)[-1])
        save_exported_media_to_file(medias_folder_path, media_file, media_id.split('/', -1)[-1])

    # Return new filled dataframe
    return df


def save_exported_media_to_file(export_dir, media_file, filename):
    """
    Write exported media item to disk at specified location with specified file name.
    Any existing file with the same name will be overwritten.
    :param export_dir:  path to directory for exports (str)
    :param media_file:  media file to write to disc (str)
    :param filename:    filename to give exported image (str)
    """
    if not os.path.exists(export_dir):
        os.makedirs(export_dir)
    file_path = os.path.join(export_dir, filename + '.jpg')
    try:
        with open(file_path, 'wb') as out_file:
            shutil.copyfileobj(media_file.raw, out_file)
            del media_file
    except:
        pass


def export_audits_csv(audit_list_json, folder_path):
    """
    Save audit CSV to folder_path
    :param audit_list_json:  list of audits in json format (list)
    :param folder_path:      path to export folder (str)
    """
    for audit in audit_list_json:
        csv_exporter = csvExporter.CsvExporter(audit, True)
        csv_export_filename = audit['audit_id']
        csv_exporter.append_converted_audit_to_bulk_export_file(
            os.path.join(folder_path, csv_export_filename + '.csv'))


def create_directories():
    """
    Creates directories for exports and medias
    :return exports_folder_path:    path to the exports folder (str)
    :return medias_folder_path:     path to the medias folder (str)
    """
    cwd = os.getcwd()
    exports_folder_path = str(cwd) + '/exports'
    medias_folder_path = str(cwd) + '/medias'
    create_directory_if_not_exists(exports_folder_path)
    create_directory_if_not_exists(medias_folder_path)
    return exports_folder_path, medias_folder_path


def create_directory_if_not_exists(path):
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


def create_dataframe():
    """
    Create empty dataframe with wanted column
    :return df:    dataframe
    """
    col_names = ['audit_id', 'audit_author', 'audit_template_name', 'audit_site' \
        , 'audit_sector', 'audit_area', 'audit_date', 'audit_score', 'audit_forecast', 'audit_month', 'audit_year',
                 'im_bonnes_pratiques_id', 'description_bonnes_pratiques', 'first_non_compliant_question']
    index = [1]
    df = pd.DataFrame(columns=col_names, index=index)
    return df


def concat_dataframes(df, df_new):
    """
    Concatenate two dataframes
    :param df:      old dataframe
    :param df_new:  new dataframe
    :return df:     dataframe concatenated
    """
    df = pd.concat([df, df_new], axis=0)
    return df


def end_dataframe(df):
    """
    Reset dataframe's index
    :param df:      dataframe
    :return df:     dataframe
    """
    df = df.iloc[1:]
    df = df.reset_index(level=0, drop=True)
    return df


def internet_on():
    """
    Check if user has access to internet
    :return bool:
    """
    conn = httplib.HTTPConnection("www.google.com", timeout=5)
    try:
        conn.request("HEAD", "/")
        conn.close()
        return True
    except:
        conn.close()
        return False


def export_actions_to_excel(actions_array, export_path=str(os.getcwd())+'/exports'):
    """
    Write Actions to 'iauditor_actions.csv' on disk at specified location
    :param export_path:     path to directory for exports
    :param actions_array:   Array of action objects to be converted to CSV and saved to disk
    """

    filename = 'iauditor_actions.csv'
    file_path = os.path.join(export_path, filename)
    if os.path.isfile(file_path):
        actions_csv = open(file_path, 'ab')
        actions_csv_wr = csv.writer(actions_csv, dialect='excel', quoting=csv.QUOTE_ALL)
    else:
        actions_csv = open(file_path, 'wb')
        actions_csv_wr = csv.writer(actions_csv, dialect='excel', quoting=csv.QUOTE_ALL)
        actions_csv_wr.writerow([
            'actionId', 'description', 'assignee', 'priority', 'priorityCode', 'status', 'statusCode', 'dueDatetime',
            'audit', 'auditId', 'linkedToItem', 'linkedToItemId', 'creatorName', 'creatorId', 'createdDatetime',
            'modifiedDatetime', 'completedDatetime', 'site', 'title'])

    for action in actions_array:
        actions_list = transform_action_object_to_list(action)
        actions_csv_wr.writerow(actions_list)
        del actions_list

    df = pd.read_csv(file_path)
    df['site'] = df['site'].str.replace(r'SLS - ', '')
    df['site'] = df['site'].str.replace(r'SLS ', '')
    df.drop_duplicates(subset=['actionId'], inplace=True)
    with pd.ExcelWriter('iAuditor_Exports.xlsx', engine='openpyxl', mode='a') as writer:
        df.to_excel(writer, sheet_name='Actions')

    print('\nActions updated in iAuditor Excel database !')


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
    actions_list.append(get_json_property(action, 'site'))
    actions_list.append(get_json_property(action, 'title'))
    return actions_list

def send_mail():
    try:
        msg = MIMEMultipart()
        msg['From'] = "tjoignant@gmail.com"
        msg['To'] = "theo.joignant@safrangroup.com"
        msg['Subject'] = str(datetime.date.today()) + " 5S Update"
        body = "Excel database update"
        msg.attach(MIMEText(body, 'plain'))
        filename = "iAuditor_Exports.xlsx"
        attachment = open("/Users/theojoignant/PycharmProjects/Safran/iAuditor_Exports.xlsx", "rb")
        p = MIMEBase('application', 'octet-stream')
        p.set_payload((attachment).read())
        encoders.encode_base64(p)
        p.add_header('Content-Disposition', "attachment; filename= %s" % filename)
        msg.attach(p)
        s = smtplib.SMTP('smtp.gmail.com', 587)
        s.starttls()
        s.login("tjoignant@gmail.com", "eqckzhihwbyhajpv")
        text = msg.as_string()
        s.sendmail("tjoignant@gmail.com", "theo.joignant@safrangroup.com", text)
        s.quit()
        print("\nEmail sent successfully")
    except:
        print("\nError when sending email")
