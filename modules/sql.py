from sqlalchemy import *
from sqlalchemy.exc import IntegrityError, OperationalError
from sqlalchemy.orm import sessionmaker

def save_exported_actions_to_db(logger, actions_array, settings, get_started):
    """
    Write Actions to 'iauditor_actions.csv' on disk at specified location
    :param get_started:
    :param logger:          the logger
    :param export_path:     path to directory for exports
    :param actions_array:   Array of action objects to be converted to CSV and saved to disk
    """
    engine = get_started[1]
    actions_db = get_started[4]

    if not actions_array:
        logger.info('No actions returned after ' + get_last_successful_actions_export(logger))
        return
    logger.info('Exporting ' + str(len(actions_array)) + ' actions')
    Session = sessionmaker(bind=engine)
    session = Session()
    bulk_actions = []
    for action in actions_array:
        action_as_list = transform_action_object_to_list(action)
        bulk_actions.append(action_as_list)
    df = pd.DataFrame.from_records(bulk_actions, columns=ACTIONS_HEADER_ROW)
    df['DatePK'] = pd.to_datetime(df['modifiedDatetime']).values.astype(np.int64) // 10 ** 6
    if settings[DB_TYPE].startswith('mysql'):
        df.replace({'DateCompleted': ''}, None, inplace=True)
        df.replace({'ConductedOn': ''}, None, inplace=True)
        df['createdDatetime'] = pd.to_datetime(df['createdDatetime'])
        df['modifiedDatetime'] = pd.to_datetime(df['modifiedDatetime'])
        df['completedDatetime'] = pd.to_datetime(df['completedDatetime'])
        df['dueDatetime'] = pd.to_datetime(df['dueDatetime'])
    df.replace({'': np.nan}, inplace=True)
    df = df.replace({np.nan: None})
    df_dict = df.to_dict(orient='records')

    try:
        session.bulk_insert_mappings(actions_db, df_dict)
    except KeyboardInterrupt:
        logger.warning('Interrupted by user, exiting.')
        session.rollback()
        sys.exit(0)
    except OperationalError as ex:
        session.rollback()
        logger.warning('Something went wrong. Here are the details: {}'.format(ex))
    except IntegrityError as ex:
        # If the bulk insert fails, we do a slower merge
        logger.warning('Unable to bulk insert (likely due to a duplicate), attempting to update')
        session.rollback()
        for action in df_dict:
            row_to_dict = actions_db(**action)
            session.merge(row_to_dict)
        logger.debug('Row successfully added/updated.')
    session.commit()

def sql_setup(logger, settings, action_or_audit):
    if settings[MERGE_ROWS] is True or False:
        merge = settings[MERGE_ROWS]
    else:
        merge = False

    if settings[ACTIONS_MERGE_ROWS] is True or False:
        actions_merge = settings[ACTIONS_MERGE_ROWS]
    else:
        actions_merge = False

    Base.metadata.clear()

    if action_or_audit == 'audit':
        if settings[SQL_TABLE] is not None:
            table = settings[SQL_TABLE]
        else:
            table = 'iauditor_data'
        Database = set_table(table, merge)
    elif action_or_audit == 'actions':
        if settings[ACTIONS_TABLE] is not None:
            table = settings[ACTIONS_TABLE]
        else:
            table = 'iauditor_actions_data'
        ActionsDatabase = set_actions_table(table, actions_merge)
    else:
        print('No Match')
        sys.exit()
    connection_string = '{}://{}:{}@{}:{}/{}'.format(settings[DB_TYPE],
                                                     settings[DB_USER],
                                                     settings[DB_PWD],
                                                     settings[DB_SERVER],
                                                     settings[DB_PORT],
                                                     settings[DB_NAME])

    engine = create_engine(connection_string)
    meta = MetaData()
    logger.debug('Making connection to ' + str(engine))
    if action_or_audit == 'audit':
        if not engine.dialect.has_table(engine, settings[SQL_TABLE], schema=settings[DB_SCHEMA]):
            logger.info(settings[SQL_TABLE] + ' not Found.')
            if settings[ALLOW_TABLE_CREATION] == 'true':
                Database.__table__.create(engine)
            elif settings[ALLOW_TABLE_CREATION] == 'false':
                logger.error('You need to create the table {} in your database before continuing. If you want the '
                             'script '
                             'to do it for you, set ALLOW_TABLE_CREATION to '
                             'True in your config file'.format(settings[SQL_TABLE]))
                sys.exit()
            else:
                validation = input('It doesn\'t look like a table called {} exists on your server. Would you like the '
                                   'script to try and create the table for you now? (If you\'re using '
                                   'docker, you need to set APPROVE_TABLE_CREATION to true in your config file) '
                                   '(y/n)  '.format(settings[SQL_TABLE]))
                validation = validation.lower()
                if validation.startswith('y'):
                    Database.__table__.create(engine)
                else:
                    logger.info('Stopping the script. Please either re-run the script or create your table manually.')
                    sys.exit()
        setup = 'complete'
        logger.info('Successfully setup Database and connection')
    else:
        if not engine.dialect.has_table(engine, settings[ACTIONS_TABLE], schema=settings[DB_SCHEMA]):
            logger.info(settings[ACTIONS_TABLE] + ' not Found.')
            if settings[ALLOW_TABLE_CREATION] == 'true':
                ActionsDatabase.__table__.create(engine)
            elif settings[ALLOW_TABLE_CREATION] == 'false':
                logger.error('You need to create the table {} in your database before continuing. If you want the '
                             'script to do it for you, set ALLOW_TABLE_CREATION to True in your '
                             'config file'.format(settings[SQL_TABLE]))
                sys.exit()
            else:
                validation = input('It doesn\'t look like a table called {} exists on your server. Would you like the '
                                   'script to try and create the table for you now? (If you\'re using '
                                   'docker, you need to set APPROVE_TABLE_CREATION to true in your config file) '
                                   '(y/n)  '.format(settings[ACTIONS_TABLE]))
                validation = validation.lower()
                if validation.startswith('y'):
                    ActionsDatabase.__table__.create(engine)
                else:
                    logger.info('Stopping the script. Please either re-run the script or create your table manually.')
                    sys.exit()
        setup = 'complete'
        logger.info('Successfully setup Database and connection')

    if action_or_audit == 'audit':
        return setup, engine, connection_string, meta, Database
    else:
        return setup, engine, connection_string, meta, ActionsDatabase


def export_audit_sql(logger, settings, audit_json, get_started):
    """
    Save audit to a database.
    :param logger:      The logger
    :param settings:    Settings from command line and configuration file
    :param audit_json:  Audit JSON
    :get_started:       Tuple containing settings
    """
    engine = get_started[1]
    database = get_started[4]

    csv_exporter = csvExporter.CsvExporter(audit_json, settings[EXPORT_INACTIVE_ITEMS_TO_CSV])
    df = csv_exporter.audit_table
    df = pd.DataFrame.from_records(df, columns=SQL_HEADER_ROW)
    df['DatePK'] = pd.to_datetime(df['DateModified']).values.astype(np.int64) // 10 ** 6
    if settings[DB_TYPE].startswith('mysql'):
        df.replace({'DateCompleted': ''}, '1900-01-01 00:00:00', inplace=True)
        df.replace({'ConductedOn': ''}, '1900-01-01 00:00:00', inplace=True)
        df['DateStarted'] = pd.to_datetime(df['DateStarted'])
        df['DateCompleted'] = pd.to_datetime(df['DateCompleted'])
        df['DateModified'] = pd.to_datetime(df['DateModified'])
        df['ConductedOn'] = pd.to_datetime(df['ConductedOn'])
    # df.replace({'': np.nan}, inplace=True)
    df.replace({'ItemScore': '', 'ItemMaxScore': '', 'ItemScorePercentage': ''}, np.nan, inplace=True)
    # df = df.replace({np.nan: None})
    df.fillna(0, inplace=True)
    df['SortingIndex'] = range(1, len(df) + 1)
    df_dict = df.to_dict(orient='records')
    Session = sessionmaker(bind=engine)
    session = Session()

    try:
        session.bulk_insert_mappings(database, df_dict)
    except KeyboardInterrupt:
        logger.warning('Interrupted by user, exiting.')
        session.rollback()
        sys.exit(0)
    except OperationalError as ex:
        session.rollback()
        logger.warning('Something went wrong. Here are the details: {}'.format(ex))
    except IntegrityError as ex:
        # If the bulk insert fails, we do a slower merge
        logger.warning('Duplicate found, attempting to update')
        session.rollback()
        for row in df_dict:
            row_to_dict = database(**row)
            session.merge(row_to_dict)
        logger.debug('Row successfully updated.')
    session.commit()