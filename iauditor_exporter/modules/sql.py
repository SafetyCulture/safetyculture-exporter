import datetime

import numpy as np
import pandas as pd
from sqlalchemy import *
from sqlalchemy.exc import IntegrityError, OperationalError
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker

import iauditor_exporter.modules.csvExporter as csvExporter
from iauditor_exporter.modules.actions import transform_action_object_to_list
from iauditor_exporter.modules.global_variables import *
from iauditor_exporter.modules.model import *


def test_sql_settings(logger, settings):
    Base = declarative_base()
    Base.metadata.clear()
    if "export_options" in settings:
        for k, v in settings["export_options"].items():
            if k in [
                "database_type",
                "database_user",
                "database_pwd",
                "database_server",
                "database_port",
                "database_name",
            ]:
                if v is None:
                    logger.warning(
                        f"All fields must be filled out to test your database settings. At least {k} is empty."
                    )
                    return False

        connection_string = "{}://{}:{}@{}:{}/{}".format(
            settings["export_options"]["database_type"],
            settings["export_options"]["database_user"],
            settings["export_options"]["database_pwd"],
            settings["export_options"]["database_server"],
            settings["export_options"]["database_port"],
            settings["export_options"]["database_name"],
        )

        engine = create_engine(connection_string, pool_pre_ping=True)
        logger.info("Attempting to connect to database...")
        try:
            conn = engine.connect()
            results = conn.execute("SELECT 1")
            logger.info("Connected successfully.")
            conn.close()
            return True
        except Exception as ex:
            logger.warning(f"Unable to connect to database: {ex}")
            return False


def sql_setup(logger, settings, action_or_audit):
    if settings[MERGE_ROWS] is True or False:
        merge = settings[MERGE_ROWS]
    else:
        merge = False

    if settings[ACTIONS_MERGE_ROWS] is True or False:
        actions_merge = settings[ACTIONS_MERGE_ROWS]
    else:
        actions_merge = False

    Base = declarative_base()
    Base.metadata.clear()

    if action_or_audit == "audit":
        if settings[SQL_TABLE] is not None:
            table = settings[SQL_TABLE]
        else:
            table = "iauditor_data"
        Database = set_table(table, merge, Base, user_schema=settings[DB_SCHEMA])
    elif action_or_audit == "actions":
        if settings[ACTIONS_TABLE] is not None:
            table = settings[ACTIONS_TABLE]
        else:
            table = "iauditor_actions_data"
        Database = set_actions_table(
            table, actions_merge, Base, user_schema=settings[DB_SCHEMA]
        )
    else:
        sys.exit()

    if HEROKU_URL in settings:
        if settings[HEROKU_URL] is not None:
            connection_string = settings[HEROKU_URL]
    else:
        connection_string = "{}://{}:{}@{}:{}/{}".format(
            settings[DB_TYPE],
            settings[DB_USER],
            settings[DB_PWD],
            settings[DB_SERVER],
            settings[DB_PORT],
            settings[DB_NAME],
        )
    engine = create_engine(connection_string)
    meta = MetaData()
    logger.debug("Making connection to " + str(engine))
    if action_or_audit == "audit":
        db_setting = settings[SQL_TABLE]
    else:
        db_setting = settings[ACTIONS_TABLE]
    if not engine.dialect.has_table(engine, db_setting, schema=settings[DB_SCHEMA]):
        logger.info(db_setting + " not Found.")
        if settings[ALLOW_TABLE_CREATION] == "true":
            Database.__table__.create(engine)
        elif settings[ALLOW_TABLE_CREATION] == "false":
            logger.error(
                "You need to create the table {} in your database before continuing. If you want the "
                "script "
                "to do it for you, set ALLOW_TABLE_CREATION to "
                "True in your config file".format(db_setting)
            )
            sys.exit()
        else:
            validation = input(
                "It doesn't look like a table called {} exists on your server. Would you like the "
                "script to try and create the table for you now?  "
                "(y/n) ".format(db_setting)
            )
            validation = validation.lower()
            if validation.startswith("y"):
                Database.__table__.create(engine)
            else:
                logger.info(
                    "Stopping the script. Please either re-run the script or create your table manually."
                )
                sys.exit()
    setup = "complete"
    logger.info("Successfully setup Database and connection")

    Session = sessionmaker(bind=engine)
    session = Session()

    return setup, session, Database


def db_formatter(settings, audit_json, all_audits=[]):
    """
    :param settings: Config settings from the config file
    :param audit_json: the audit in JSON format
    :param all_audits: a list that we append processed audits too
    :return: An updated list

    This function takes the audit json and converts it into a table. It then does various conversions to the data
    to get it ready for insertion into the database. DatePK acts as part of the primary key to detect duplicates.
    Some databases handle dates differently so there is additional handling here. Similarly with empty integer and
    string values, as some databases reject None values.

    If you are editing the code to work with a new database type, this is likely where you want to do your edits.
    SQLAlchemy should manage the actual access, it's just ensuring the data is correctly formatted that matters.

    """

    csv_exporter = csvExporter.CsvExporter(
        audit_json, settings[EXPORT_INACTIVE_ITEMS_TO_CSV]
    )
    df = csv_exporter.audit_table
    df = pd.DataFrame.from_records(df, columns=SQL_HEADER_ROW)
    df["Comment"] = df["Comment"].str.slice(0, 60000)
    df["DatePK"] = pd.to_datetime(df["DateModified"]).values.astype(np.int64) // 10 ** 6
    empty_score = None
    if settings[DB_TYPE].startswith("postgres"):
        empty_score = np.nan
        df.replace(
            {"ItemScore": "", "ItemMaxScore": "", "ItemScorePercentage": "",},
            empty_score,
            inplace=True,
        )
        df.replace({"DateCompleted": ""}, np.datetime64(None), inplace=True)
        df.replace({"ConductedOn": ""}, np.datetime64(None), inplace=True)
    elif settings[DB_TYPE].startswith("mysql"):
        df["DateStarted"] = pd.to_datetime(df["DateStarted"])
        df["DateCompleted"] = pd.to_datetime(df["DateCompleted"])
        df["DateModified"] = pd.to_datetime(df["DateModified"])
        df["ConductedOn"] = pd.to_datetime(df["ConductedOn"])
        empty_score = 0.0
        df.replace(
            {"ItemScore": "", "ItemMaxScore": "", "ItemScorePercentage": "",},
            empty_score,
            inplace=True,
        )
        df.replace(
            {"ConductedOn": pd.NaT, "DateCompleted": pd.NaT},
            "1970-01-01T00:00:01",
            inplace=True,
        )
    df.replace(r"^\s*$", None, regex=True, inplace=True)
    df["SortingIndex"] = range(1, len(df) + 1)
    df.replace(np.nan, None, inplace=True)
    df.replace({"DateCompleted": ""}, None, inplace=True)
    df.replace({"ConductedOn": ""}, None, inplace=True)

    df_dict = df.to_dict(orient="records")
    all_audits.append(df_dict)


# def export_audit_sql(logger, settings, audit_json, get_started):
#     """
#     Save audit to a database.
#     :param logger:      The logger
#     :param settings:    Settings from command line and configuration file
#     :param audit_json:  Audit JSON
#     :get_started:       Tuple containing settings
#     """
#     database = get_started[2]
#     session = get_started[1]
#
#     csv_exporter = csvExporter.CsvExporter(
#         audit_json, settings[EXPORT_INACTIVE_ITEMS_TO_CSV]
#     )
#     df = csv_exporter.audit_table
#     df = pd.DataFrame.from_records(df, columns=SQL_HEADER_ROW)
#     df["DatePK"] = pd.to_datetime(df["DateModified"]).values.astype(np.int64) // 10 ** 6
#     if settings[DB_TYPE].startswith("postgres"):
#         df.replace({"DateCompleted": ""}, np.datetime64(None), inplace=True)
#         df.replace({"ConductedOn": ""}, np.datetime64(None), inplace=True)
#         empty_value = np.nan
#         empty_score = empty_value
#     elif settings[DB_TYPE].startswith("mysql"):
#         df.replace(
#             {"ItemScore": "", "ItemMaxScore": "", "ItemScorePercentage": ""},
#             0.0,
#             inplace=True,
#         )
#         empty_value = "1970-01-01T00:00:01"
#         df.replace({"DateCompleted": ""}, empty_value, inplace=True)
#         df.replace({"ConductedOn": ""}, empty_value, inplace=True)
#         df["DateStarted"] = pd.to_datetime(df["DateStarted"])
#         df["DateCompleted"] = pd.to_datetime(df["DateCompleted"])
#         df["DateModified"] = pd.to_datetime(df["DateModified"])
#         df["ConductedOn"] = pd.to_datetime(
#             df["ConductedOn"], format="%Y-%m-%d %H:%M:%S", utc=False
#         )
#         df["ConductedOn"] = df["ConductedOn"].dt.tz_localize(None)
#         empty_value = None
#         empty_score = 0.0
#     else:
#         empty_value = None
#         empty_score = empty_value
#
#     df.replace(
#         {"ItemScore": "", "ItemMaxScore": "", "ItemScorePercentage": ""},
#         empty_score,
#         inplace=True,
#     )
#     df.replace(r"^\s*$", empty_value, regex=True, inplace=True)
#     df["SortingIndex"] = range(1, len(df) + 1)
#     df_dict = df.to_dict(orient="records")
#     try:
#         session.bulk_insert_mappings(database, df_dict)
#     except KeyboardInterrupt:
#         logger.warning("Interrupted by user, exiting.")
#         session.rollback()
#         session.close()
#         sys.exit(0)
#     except OperationalError as ex:
#         session.rollback()
#         session.close()
#         logger.warning("Something went wrong. Here are the details: {}".format(ex))
#     except IntegrityError as ex:
#         # If the bulk insert fails (likely due to a duplicate), we do a slower merge
#         logger.warning("Duplicate found, attempting to update")
#         session.rollback()
#         for row in df_dict:
#             row_to_dict = database(**row)
#             session.merge(row_to_dict)
#         logger.debug("Row successfully updated.")
#     session.commit()


def bulk_import_sql(logger, df_dict, get_started):
    """
    Save audit to a database.
    :param logger:      The logger
    :param settings:    Settings from command line and configuration file
    :param audit_json:  Audit JSON
    :get_started:       Tuple containing settings
    """
    database = get_started[2]
    session = get_started[1]

    # SQLAlchemy allows us to bulk insert or update data by giving it a list of dictionaries. Inserts are faster so
    # we always try this first. If this fails, it's almost always due to a duplicate row. In this instance, we pick up
    # the exception and do a bulk update instead. If all else fails we'll rollback and give an error.

    for audit in df_dict:
        try:
            logger.info(f"Inserting {audit[0]['AuditID']} into the database")
            session.bulk_insert_mappings(database, audit, render_nulls=True)
        except KeyboardInterrupt:
            logger.warning("Interrupted by user, exiting.")
            session.rollback()
            session.close()
            sys.exit(0)
        except IntegrityError as ex:
            # If the bulk insert fails, we do a slower merge
            logger.info("Duplicate found, attempting to update")
            session.rollback()
            for row in audit:
                row_to_dict = database(**row)
                session.merge(row_to_dict)
            logger.debug("Row successfully updated.")
        except Exception as ex:
            session.rollback()
            session.close()
            logger.warning("Something went wrong. Here are the details: {}".format(ex))
            sys.exit()
        session.commit()


def end_session(session):
    session.close()


def query_max_last_modified(session, database):
    qry = session.query(func.max(database.DateModified).label("max"))
    res = qry.one()
    if not res.max:
        return "2000-01-01T00:00:00.000Z"
    else:
        new_last_successful = res.max + datetime.timedelta(0, 10)
        new_last_successful = str(new_last_successful)
        return new_last_successful


def save_exported_actions_to_db(logger, actions_array, settings, get_started):
    """
    Write Actions to 'iauditor_actions.csv' on disk at specified location
    :param get_started:
    :param logger:          the logger
    :param export_path:     path to directory for exports
    :param actions_array:   Array of action objects to be converted to CSV and saved to disk
    """

    actions_db = get_started[2]
    session = get_started[1]

    if not actions_array:
        logger.info("No actions returned.")
        return
    logger.info("Exporting " + str(len(actions_array)) + " actions")
    bulk_actions = []
    for action in actions_array:
        action_as_list = transform_action_object_to_list(action)
        bulk_actions.append(action_as_list)
    df = pd.DataFrame.from_records(bulk_actions, columns=ACTIONS_HEADER_ROW)
    df["DatePK"] = (
        pd.to_datetime(df["modifiedDatetime"]).values.astype(np.int64) // 10 ** 6
    )
    if settings[DB_TYPE].startswith(("mysql", "postgres")):
        df.replace({"DateCompleted": ""}, None, inplace=True)
        df.replace({"ConductedOn": ""}, None, inplace=True)
        df["createdDatetime"] = pd.to_datetime(df["createdDatetime"])
        df["modifiedDatetime"] = pd.to_datetime(df["modifiedDatetime"])
        df["completedDatetime"] = pd.to_datetime(df["completedDatetime"])
        df["dueDatetime"] = pd.to_datetime(df["dueDatetime"])
    df.replace({"": np.nan}, inplace=True)
    df = df.replace({np.nan: None})
    df_dict = df.to_dict(orient="records")

    try:
        session.bulk_insert_mappings(actions_db, df_dict, render_nulls=True)
    except KeyboardInterrupt:
        logger.warning("Interrupted by user, exiting.")
        session.rollback()
        sys.exit(0)
    except OperationalError as ex:
        session.rollback()
        logger.warning("Something went wrong. Here are the details: {}".format(ex))
    except IntegrityError as ex:
        # If the bulk insert fails, we do a slower merge
        logger.warning(
            "Unable to bulk insert (likely due to a duplicate), attempting to update"
        )
        session.rollback()
        for action in df_dict:
            row_to_dict = actions_db(**action)
            session.merge(row_to_dict)
        logger.debug("Row successfully added/updated.")
    session.commit()
