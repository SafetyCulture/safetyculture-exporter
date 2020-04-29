from sqlalchemy import Column, String, Integer, Float, DateTime, Boolean, BigInteger, Text
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


def set_table(table, merge):
    class Database(Base):
        __tablename__ = table
        SortingIndex = Column(Integer)
        ItemType = Column(String(20))
        Label = Column(Text())
        Response = Column(Text())
        Comment = Column(Text())
        MediaHypertextReference = Column(Text())
        Latitude = Column(String(50))
        Longitude = Column(String(50))
        ItemScore = Column(Float)
        ItemMaxScore = Column(Float)
        ItemScorePercentage = Column(Float)
        Mandatory = Column(Boolean)
        FailedResponse = Column(Boolean)
        Inactive = Column(Boolean)
        AuditID = Column(String(100), primary_key=True, autoincrement=False)
        ItemID = Column(String(100), primary_key=True, autoincrement=False)
        if merge is False:
            DatePK = Column(String(20), primary_key=True, autoincrement=False)
        else:
            DatePK = Column(String(20))
        ResponseID = Column(Text())
        ParentID = Column(String(100))
        AuditOwner = Column(Text())
        AuditAuthor = Column(Text())
        AuditOwnerID = Column(Text())
        AuditAuthorID = Column(String(100))
        AuditName = Column(Text())
        AuditScore = Column(Float)
        AuditMaxScore = Column(Float)
        AuditScorePercentage = Column(Float)
        AuditDuration = Column(Float)
        DateStarted = Column(DateTime)
        DateCompleted = Column(DateTime)
        DateModified = Column(DateTime)
        TemplateID = Column(String(100))
        TemplateName = Column(Text())
        TemplateAuthor = Column(Text())
        TemplateAuthorID = Column(String(100))
        ItemCategory = Column(Text())
        RepeatingSectionParentID = Column(String(100))
        DocumentNo = Column(Text())
        ConductedOn = Column(DateTime)
        PreparedBy = Column(Text())
        Location = Column(Text())
        Personnel = Column(Text())
        ClientSite = Column(Text())
        AuditSite = Column(Text())
        AuditArea = Column(Text())
        AuditRegion = Column(Text())
        Archived = Column(Boolean)
    return Database


SQL_HEADER_ROW = [
    'SortingIndex',
    'ItemType',
    'Label',
    'Response',
    'Comment',
    'MediaHypertextReference',
    'Latitude',
    'Longitude',
    'ItemScore',
    'ItemMaxScore',
    'ItemScorePercentage',
    'Mandatory',
    'FailedResponse',
    'Inactive',
    'ItemID',
    'ResponseID',
    'ParentID',
    'AuditOwner',
    'AuditAuthor',
    'AuditOwnerID',
    'AuditAuthorID',
    'AuditName',
    'AuditScore',
    'AuditMaxScore',
    'AuditScorePercentage',
    'AuditDuration',
    'DateStarted',
    'DateCompleted',
    'DateModified',
    'AuditID',
    'TemplateID',
    'TemplateName',
    'TemplateAuthor',
    'TemplateAuthorID',
    'ItemCategory',
    'RepeatingSectionParentID',
    'DocumentNo',
    'ConductedOn',
    'PreparedBy',
    'Location',
    'Personnel',
    'ClientSite',
    'AuditSite',
    'AuditArea',
    'AuditRegion',
    'Archived'
]


def set_actions_table(table, merge):
    class ActionsDatabase(Base):
        __tablename__ = table
        id = Column(Integer, primary_key=False, autoincrement=True)
        description = Column(Text())
        assignee = Column(Text())
        priority = Column(Text())
        priorityCode = Column(Integer)
        status = Column(String(20))
        statusCode = Column(Integer)
        dueDatetime = Column(DateTime)
        actionId = Column(String(100), primary_key=True, autoincrement=False)
        if merge is False:
            DatePK = Column(BigInteger, autoincrement=False)
        else:
            DatePK = Column(BigInteger, primary_key=True, autoincrement=False)
        audit = Column(Text())
        auditId = Column(String(50))
        linkedToItem = Column(Text())
        linkedToItemId = Column(Text())
        creatorName = Column(Text())
        creatorId = Column(String(50))
        createdDatetime = Column(DateTime)
        modifiedDatetime = Column(DateTime)
        completedDatetime = Column(DateTime)
    return ActionsDatabase


ACTIONS_HEADER_ROW = [
    'actionId',
    'description',
    'assignee',
    'priority',
    'priorityCode',
    'status',
    'statusCode',
    'dueDatetime',
    'audit',
    'auditId',
    'linkedToItem',
    'linkedToItemId',
    'creatorName',
    'creatorId',
    'createdDatetime',
    'modifiedDatetime',
    'completedDatetime'
]