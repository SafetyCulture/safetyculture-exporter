# Columns

The exported data is split down into many columns which can appear confusing at first. Keep in mind that the data from this tool is designed to be pivoted so don't be put off if opening the files in Excel appears unusual. 

!!! Info
    The columns are listed here in the same order they appear in the output. All columns are of type `String` in SQL unless stated otherwise. 

### SortingIndex
**SQL Data Type:** `Integer`

A numerical index to aid in sorting inspections into the order they're conducted. The number runs from 1 up until the last question of the inspection then resets for the next inspection. 

!!! Tip
    In order to use this column for sorting, you need to first sort by AuditID, then by SortingIndex. 

### ItemType
This column describes the type of item information in this particular row. Possible values are:
#### **Question Types**
* `question`
* `list`
* `media`
* `signature`
* `slider`
* `text`
* `textsingle`
* `address`
* `checkbox`
* `datetime`
* `drawing`
#### **Organisational Elements** 
* `category`
* `section`
* `smartfield`
* `information`
#### **Repeating Section-Related Elements**
* `dynamicfield`
* `element`
* `primelement`

!!! Tip
    In some cases, the type may change between question and list. This occurs if the response set associated with a question changes. A question becomes a list type under these circumstances:
    
    * More than 5 possible response options
    * One response is more than 30 characters in length


!!! Tip
    In some cases, the type may change between text and textsingle. This occurs if the text field is changed from a single line to a paragraph. 

### Label
The question itself asked within the given inspection. 

!!! Tip
    If a question is reworded, any newly completed inspections will have the new label so be careful filtering on the label.

### Response
The given by the person conducting the audit. 

### Comment
iAuditor allows users to add notes to questions. If any notes have been left on a question, they'll be logged here

!!! Tip
    These are essentially free text fields so users can type whatever they want. Visualisations like Word Clouds may be a good choice with these, or picking out key words.

### MediaHyperTextReference
iAuditor allows users to add images to almost all question types. If one or more images have been attached to a question, they will be logged in this column. All images in iAuditor are given a unique ID which we call a `media_id`. When logged in this column, we use `media_id` and append the file extension. 

!!! Tip
    Usually the image format is .jpg however we support other image formats. It's also possible this could change over time, hence why it's logged here rather than assumed.
    
    This exporter tool allows for the exporting of media. When you do this, the images are stored in the same format logged here (`media_id` + file extension) making it easy to reference images you've downloaded. 
    
    If you want to make API calls to the `media_id`, you'll need to split the column first and remove the file extension. 

### Latitude
Some questions in iAuditor, such as the `location` and `address` item types can have location information attached to them. If that information is available, it is logged here.

### Longitude 
See Latitude

### ItemScore
**SQL Data Type**: `Float`

If a question has been scored (this is set when creating a template), its score is logged here. 

!!! Tip
    Item scores can be decimals so consider this column a `Float`. 

### ItemMaxScore
**SQL Data Type:** `Float`

The maximum score a question can potentially score. 

### ItemScorePercentage
**SQL Data Type:** `Float`

The percentage score of a question. 

!!! Tip
    The percentage is represented in full, so 100% is shown as 100. Depending on your tool, you may wish to divide this column by 100 so it renders correctly as a percentage. (this is the case in PowerBI)

### Mandatory
**SQL Data Type:** `Boolean`

In the iAuditor template editor, questions can be marked as mandatory (e.g.  an inspection cannot be marked as complete until all mandatory questions have been answered.) All items are given either a True or False value. 

### FailedResponse
**SQL Data Type:** `Boolean`

Particular responses in iAuditor can be given a _failed_ status. This means that the particular response was considered to be something incorrect and tends to be an important data point to review. 

### Inactive
**SQL Data Type:** `Boolean`

Sometimes questions within inspections may never be seen by the user. For example, we have logic fields where questions appear depending on the response of a previous question. In these instances, the questions that were not activated by the logic field would be considered inactive. 

By default, the exporter does not export inactive items so usually this is False for everything. 

### AuditID
**SQL Note:** Forms part of the Primary Key in SQL

This is the unique ID of a particular audit. 

!!! Tip
    Unique across the entire iAuditor platform - not just your own organisation.
    
    Never changes. 

### ItemID
**SQL Note:** Forms part of the Primary Key in SQL

Every item in a particular inspection is allocated an Item ID. 

!!! Tip
    Item IDs are _only unique within an inspection_. The ID is inherited from the parent template, so questions will have the same Item ID across audits conducted from the same template. 

### DatePK
**SQL Note:** Forms part of the Primary Key in SQL

Only present in the SQL output, not required or included for the CSV export. 

The date presented in EPOCH format (e.g. number of seconds since January 1st 1970) - Exists to be used as part of three column primary key with AuditID and ItemID. Please see the Primary Key section for further explanation. 

### ResponseID
Any response that isn't free text will be assigned an ID. These are unique _within an audit_. If you want to track a particular response, tracking the ID is usually more robust as it avoids any issues should someone rename a response. 

### ParentID
All items (except the very top item) have a parent ID which signifies the item that the current item proceeds. Traversing the ItemID > ParentID chain can help make sense of how a particular inspection is conducted. 

### AuditOwner
The name of the user who originally started the inspection

### AuditOwnerID
As above, but returns the unique User ID of the particular user. 

### AuditAuthor
The name of the user who last modified the inspection

### AuditAuthorID
As above, but returns the unique User ID of the particular user. 

### AuditName
The title of the inspection as generated in iAuditor

### AuditScore
**SQL Data Type:** `Float`

The score of the entire inspection

### AuditMaxScore
**SQL Data Type:** `Float`

The maximum score the particular inspection can achieve

### AuditScorePercentage
**SQL Data Type:** `Float`

The percentage score of the given inspection. 

!!! Tip
    The percentage is represented in full, so 100% is shown as 100. Depending on your tool, you may wish to divide this column by 100 so it renders correctly as a percentage. (This is the case in PowerBI)

### AuditDuration
**SQL Data Type:** `Integer`

The time taken to conduct an inspection, recorded in seconds. 

!!! Tip
    If a user conducts an inspection using our website, the duration isn't currently logged and will return blank. 

### DateStarted
**SQL Data Type:** `DateTime`

The date the inspection was started presented in UTC format. 

!!! Tip
    This is the date when the inspection was created on our platform. In the majority of cases, this will be the same or very close to the 'ConductedOn' date, however sometimes it may appear wildly different. If an inspection is started (most likely via our API), then subsequently edited and completed by a user, the DateStarted may appear anomalous. Where possible, use `ConductedOn` instead. 

### DateModified
**SQL Data Type:** `DateTime`

The date the inspection was last modified presented in UTC format. 

!!! Tip
    The date modified will update whenever an inspection is modified by a user, so it's possible that this may change after an inspection has been marked as complete. 

### DateCompleted
**SQL Data Type:** `DateTime`

The date the inspection was last marked as complete presented in UTC format. 

!!! Tip
    The date completed will update if 'Complete' is pressed in the inspection a second time, usually after modifications have been made. 

### TemplateID
The unique ID of the template.

!!! Tip
    The Template ID is unique across the entire iAuditor platform
    
    The Template ID never changes

### TemplateName
The name of the template at the time the inspection was completed

### TemplateAuthor
The user who created the template

### TemplateAuthorID
As above, but the unique User ID is returned instead.

### ItemCategory
Within a template, questions can be nested inside `sections` and `categorys`. For each item in the inspection, we traverse the ItemID and ParentID chain until we find a item type of either `section` or `category`. When this occurs, the `label` of that item is logged here. 

!!! Tip
    The ItemCategory column is incredibly useful for grouping data together with like questions. 
    
    You may need to work with those creating templates to ensure questions are nested effectively to give the required ItemCategory

### RepeatingSectionParentID
Within iAuditor, it is possible to have sets of questions that users can chose to repeat. For example, if you're inspecting rooms in a hotel, you may use a repeating section to repeat the same set of questions for each room within the hotel. To make modelling easier, this column contains the ID of the item that all questions within a repeating section have in common. If you're working with repeating sections (they're of type `dynamicfield`), you'll want to group the data on this column. If a question isn't part of a repeating section, this column will be blank. 

### DocumentNo
This is an optional feature of iAuditor where a document number is generated. This is usually different in each template. 

!!! Tip
    Document Numbers will have an incrementing number as part of the resulting value. However, this is created by a users device so it's possible that multiple inspections could have the same document number. For this reason, it tends to be less useful to model on the Document Number.

### ConductedOn
**SQL Data Type:** `DateTime`

Assuming the template includes the 'Conducted On' question in its title page (if you're reviewing the JSON, this is referred to as the 'header'), this will be populated with a UTC time stamp. As this is a question, it's possible for the user to pick any time value they like, however usually they'll tap 'Now' and it'll insert the current time. 

### PreparedBy
Assuming the template includes the 'Prepared By' question in its title page (if you're reviewing the JSON, this is referred to as the 'header'), this will be populated with the name of the person who conducted the inspection.

### Location
Assuming the template includes the 'Location' question in its title page (if you're reviewing the JSON, this is referred to as the 'header'), this will be populated with the inputted location. The formatting of this can vary quite a bit, for better location tracking use the `Latitude` and `Longitude` columns. 

### Personnel
This is an old field from an older version of our template editor. We've now removed this as an option so it'll only be populated if you're accessing data from an organisation created before late-2019. The field was a free text field so its input will vary depending on use case. 

### Clientsite
This is an old field from an older version of our template editor. We've now removed this as an option so it'll only be populated if you're accessing data from an organisation created before late-2019. The field was a free text field so its input will vary depending on use case. 

### AuditSite
The selected Site (specifically referring to the iAuditor Sites feature)

!!! Tip
    Sites are centrally managed by an admin and so make for a great way to group data together.
    
    Sites can be renamed so they may change in the data if this occurs. However, this can only be done by an admin. 

### Audit Area
As above, but displays the area information instead. 

### Audit Region
As above, but displays the region information instead. 

### Archived
**SQL Data Type:** `Boolean`

The archived status of a given inspection. You are able to download everything from your account including archived inspections if you want. If this config option is enabled, anything downloaded from the archive will be marked as True. 

!!! Tip
    Archiving an audit causes the modified date to update, as such the tool will re-download anything moved to the archive along with its new archived status. 

