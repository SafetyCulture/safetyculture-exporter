# The Database Model

As part of the SQL support in this tool, if the table doesn't already exist, the script will attempt to create it for you. This document will explain the structure of the database, and how we manage Primary Keys. 


As discussed [here](../../understanding-the-data/datastructure/#auditid), the `AuditID` is always unique for a particular inspection, and the `ItemID` is always unique for a particular item within an inspection. This allows us to reference a item uniquely within the dataset. However, this is making the assumption that an inspection never reappears which isn't necessarily true. The exporter tool (and the API) finds inspections based on the date they were last modified (recorded as the `DateModified` in the dataset) so if an inspection is modified after first being recorded, we'll see the same inspection reappear. 

Due to this, the database can be configured to manage this in different ways. 

## Primary Keys

In a lot of databases, the Primary Key would be a single column however when working with iAuditor, we combine either two or three columns to obtain a unique reference. By defining what makes a column unique, we can also use it as a means to update records that already exist. 

### If merge_rows/actions_merge_rows is disabled
When this is set to `false` in the config file, the Primary Key of the database is formed using three columns:

* `AuditID`
* `ItemID`
* `DatePK`

We've already discussed how the `AuditID` and `ItemID` can form a unique reference, but we need to allow for the possibility of a modified inspection reappearing. This is done using the column `DatePK`. 

`DatePK` is a custom column added by the script (it isn't present in the usual API output), and is the EPOCH representation (the number of seconds since January 1st 1970) of `DateModified`. It isn't possible to use a column of type `DateTime` as part of a Primary Key, but we can use a number (a Big Integer in this case) hence the creation of `DatePK`.

By combining these three columns, we can guarantee that all rows are truly unique, no matter when they were modified. 

!!! Tip
    When using this method, you'll need to use a SQL statement to only bring in the most recently modified value.

### If merge_rows/actions_merge_rows is enabled
When this is set to `true` in the config file, the Primary Key of the database is formed using two columns:

* `AuditID`
* `ItemID`

!!! Info
    The exporting of inactive items is also forced to `true`. See the warning below for some discussion around this.
    
When enabled, we don't use `DatePK` within our Primary Key, instead relying purely on the `AuditID` and `ItemID` to look for unique columns. The script attempts to bulk upload each inspection to the database. If it isn't able to bulk upload the inspection due to a violation of the Primary Key (e.g. there's a duplicate), it goes through each row within the inspection, searches for the `AuditID`+`ItemID` row in question and _updates_ the row instead.  

!!! Warning
    This setting may seem like the more obvious choice however it has some important caveats:
    
    * If you get a lot of inspections coming back through (particular prevalent if you've opted to download incomplete inspections as they may be downloaded multiple times during an inspection process), the tool will run slower as it needs to select the duplicate row and update it. The bulk upload that occurs when there is no duplicate is _significantly_ faster. It also has the potential to put additional strain on your SQL server. 
    
    * The exporting of inactive items is forced to `true` as to ensure every item in the inspection is updated should a change be made. For example, let's say you have an inspection with a logic field which displays 5 questions if you select 'Yes' and 3 questions if you select 'No'. Without enabling the export of inactive items, if the first time we see the inspection, 'Yes' has been selected, the other 3 questions wouldn't be logged at all. This isn't necessarily a problem, but if the inspection is then modified to be 'No', the 5 questions under 'Yes' wouldn't be marked as inactive in the dataset (because they'd be excluded), but the 3 questions under 'No' would still be introduced. In this scenario, the data is rendered incorrect as the questions under 'Yes' would be considered answered in our dataset when in fact they're no longer present. 
    
    * Due to inactive items being exported, your dataset has the potential to be bigger, especially if your templates make heavy use of logic fields. For some templates this could easily double or triple the number of rows created. 
    * If you need advice on this setting, please get in touch.