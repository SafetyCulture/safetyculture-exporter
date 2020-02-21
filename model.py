from sqlalchemy import Column, String, Integer, Float, DateTime, Boolean, BigInteger
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


def set_table(table, merge):
    class Database(Base):
        if merge is False:
            __tablename__ = table
            SortingIndex = Column(Integer)
            ItemType = Column(String(20))
            Label = Column(String(None))
            Response = Column(String(None))
            Comment = Column(String(None))
            MediaHypertextReference = Column(String(None))
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
            DatePK = Column(String(20), primary_key=True, autoincrement=False)
            ResponseID = Column(String(100))
            ParentID = Column(String(100))
            AuditOwner = Column(String(None))
            AuditAuthor = Column(String(None))
            AuditOwnerID = Column(String(None))
            AuditAuthorID = Column(String(100))
            AuditName = Column(String(None))
            AuditScore = Column(Float)
            AuditMaxScore = Column(Float)
            AuditScorePercentage = Column(Float)
            AuditDuration = Column(Float)
            DateStarted = Column(DateTime)
            DateCompleted = Column(DateTime)
            DateModified = Column(DateTime)
            TemplateID = Column(String(100))
            TemplateName = Column(String(None))
            TemplateAuthor = Column(String(None))
            TemplateAuthorID = Column(String(100))
            ItemCategory = Column(String(None))
            RepeatingSectionParentID = Column(String(100))
            DocumentNo = Column(String(None))
            ConductedOn = Column(DateTime)
            PreparedBy = Column(String(None))
            Location = Column(String(None))
            Personnel = Column(String(None))
            ClientSite = Column(String(None))
            AuditSite = Column(String(None))
            AuditArea = Column(String(None))
            AuditRegion = Column(String(None))
            Archived = Column(Boolean)
        if merge is True:
            __tablename__ = table
            SortingIndex = Column(Integer)
            ItemType = Column(String(20))
            Label = Column(String(None))
            Response = Column(String(None))
            Comment = Column(String(None))
            MediaHypertextReference = Column(String(None))
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
            DatePK = Column(String(20))
            ResponseID = Column(String(100))
            ParentID = Column(String(100))
            AuditOwner = Column(String(None))
            AuditAuthor = Column(String(None))
            AuditOwnerID = Column(String(None))
            AuditAuthorID = Column(String(100))
            AuditName = Column(String(None))
            AuditScore = Column(Float)
            AuditMaxScore = Column(Float)
            AuditScorePercentage = Column(Float)
            AuditDuration = Column(Float)
            DateStarted = Column(DateTime)
            DateCompleted = Column(DateTime)
            DateModified = Column(DateTime)
            TemplateID = Column(String(100))
            TemplateName = Column(String(None))
            TemplateAuthor = Column(String(None))
            TemplateAuthorID = Column(String(100))
            ItemCategory = Column(String(None))
            RepeatingSectionParentID = Column(String(100))
            DocumentNo = Column(String(None))
            ConductedOn = Column(DateTime)
            PreparedBy = Column(String(None))
            Location = Column(String(None))
            Personnel = Column(String(None))
            ClientSite = Column(String(None))
            AuditSite = Column(String(None))
            AuditArea = Column(String(None))
            AuditRegion = Column(String(None))
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
        if merge is False:
            __tablename__ = table
            id = Column(Integer, primary_key=False, autoincrement=True)
            description = Column(String(None))
            assignee = Column(String(None))
            priority = Column(String(None))
            priorityCode = Column(Integer)
            status = Column(String(20))
            statusCode = Column(Integer)
            dueDatetime = Column(DateTime)
            actionId = Column(String(100), primary_key=True, autoincrement=False)
            DatePK = Column(BigInteger, autoincrement=False)
            audit = Column(String(None))
            auditId = Column(String(50))
            linkedToItem = Column(String(None))
            linkedToItemId = Column(String(50))
            creatorName = Column(String(None))
            creatorId = Column(String(50))
            createdDatetime = Column(DateTime)
            modifiedDatetime = Column(DateTime)
            completedDatetime = Column(DateTime)
        else:
            __tablename__ = table
            id = Column(Integer, primary_key=False, autoincrement=True)
            description = Column(String(None))
            assignee = Column(String(None))
            priority = Column(String(None))
            priorityCode = Column(Integer)
            status = Column(String(20))
            statusCode = Column(Integer)
            dueDatetime = Column(DateTime)
            actionId = Column(String(100), primary_key=True, autoincrement=False)
            DatePK = Column(BigInteger, primary_key=True, autoincrement=False)
            audit = Column(String(None))
            auditId = Column(String(50))
            linkedToItem = Column(String(None))
            linkedToItemId = Column(String(50))
            creatorName = Column(String(None))
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