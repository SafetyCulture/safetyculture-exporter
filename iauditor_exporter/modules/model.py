from sqlalchemy import (
    Column,
    String,
    Integer,
    Float,
    DateTime,
    Boolean,
    BigInteger,
    Numeric,
    UnicodeText,
    schema,
)


def set_table(table, merge, Base, user_schema=None):
    class Database(Base):
        __tablename__ = table
        if user_schema:
            __table_args__ = {"schema": user_schema}
        SortingIndex = Column(Integer)
        ItemType = Column(String(20))
        Label = Column(UnicodeText())
        Response = Column(UnicodeText())
        Comment = Column(UnicodeText())
        MediaHypertextReference = Column(UnicodeText())
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
            DatePK = Column(BigInteger, primary_key=True, autoincrement=False)
        else:
            DatePK = Column(BigInteger)
        ResponseID = Column(UnicodeText())
        ParentID = Column(String(100))
        AuditOwner = Column(UnicodeText())
        AuditAuthor = Column(UnicodeText())
        AuditOwnerID = Column(UnicodeText())
        AuditAuthorID = Column(String(100))
        AuditName = Column(UnicodeText())
        AuditScore = Column(Float)
        AuditMaxScore = Column(Float)
        AuditScorePercentage = Column(Float)
        AuditDuration = Column(Float)
        DateStarted = Column(DateTime)
        DateCompleted = Column(DateTime)
        DateModified = Column(DateTime)
        TemplateID = Column(String(100))
        TemplateName = Column(UnicodeText())
        TemplateAuthor = Column(UnicodeText())
        TemplateAuthorID = Column(String(100))
        ItemCategory = Column(UnicodeText())
        RepeatingSectionParentID = Column(String(100))
        DocumentNo = Column(UnicodeText())
        ConductedOn = Column(DateTime)
        PreparedBy = Column(UnicodeText())
        Location = Column(UnicodeText())
        Personnel = Column(UnicodeText())
        ClientSite = Column(UnicodeText())
        AuditSite = Column(UnicodeText())
        AuditArea = Column(UnicodeText())
        AuditRegion = Column(UnicodeText())
        Archived = Column(Boolean)
        if user_schema:
            schema = user_schema

    return Database


SQL_HEADER_ROW = [
    "SortingIndex",
    "ItemType",
    "Label",
    "Response",
    "Comment",
    "MediaHypertextReference",
    "Latitude",
    "Longitude",
    "ItemScore",
    "ItemMaxScore",
    "ItemScorePercentage",
    "Mandatory",
    "FailedResponse",
    "Inactive",
    "ItemID",
    "ResponseID",
    "ParentID",
    "AuditOwner",
    "AuditAuthor",
    "AuditOwnerID",
    "AuditAuthorID",
    "AuditName",
    "AuditScore",
    "AuditMaxScore",
    "AuditScorePercentage",
    "AuditDuration",
    "DateStarted",
    "DateCompleted",
    "DateModified",
    "AuditID",
    "TemplateID",
    "TemplateName",
    "TemplateAuthor",
    "TemplateAuthorID",
    "ItemCategory",
    "RepeatingSectionParentID",
    "DocumentNo",
    "ConductedOn",
    "PreparedBy",
    "Location",
    "Personnel",
    "ClientSite",
    "AuditSite",
    "AuditArea",
    "AuditRegion",
    "Archived",
]


def set_actions_table(table, merge, Base, user_schema=None):
    class ActionsDatabase(Base):
        __tablename__ = table
        if user_schema:
            __table_args__ = {"schema": user_schema}
        id = Column(Integer, primary_key=False, autoincrement=True)
        title = Column(UnicodeText())
        description = Column(UnicodeText())
        site = Column(UnicodeText())
        assignee = Column(UnicodeText())
        priority = Column(UnicodeText())
        priorityCode = Column(Integer)
        status = Column(String(20))
        statusCode = Column(Integer)
        dueDatetime = Column(DateTime)
        actionId = Column(String(100), primary_key=True, autoincrement=False)
        if merge is False:
            DatePK = Column(BigInteger, autoincrement=False)
        else:
            DatePK = Column(BigInteger, primary_key=True, autoincrement=False)
        audit = Column(UnicodeText())
        auditId = Column(String(50))
        linkedToItem = Column(UnicodeText())
        linkedToItemId = Column(UnicodeText())
        creatorName = Column(UnicodeText())
        creatorId = Column(String(50))
        createdDatetime = Column(DateTime)
        modifiedDatetime = Column(DateTime)
        completedDatetime = Column(DateTime)
        if schema:
            schema = user_schema

    return ActionsDatabase


ACTIONS_HEADER_ROW = [
    "actionId",
    "title",
    "description",
    "site",
    "assignee",
    "priority",
    "priorityCode",
    "status",
    "statusCode",
    "dueDatetime",
    "audit",
    "auditId",
    "linkedToItem",
    "linkedToItemId",
    "creatorName",
    "creatorId",
    "createdDatetime",
    "modifiedDatetime",
    "completedDatetime",
]
