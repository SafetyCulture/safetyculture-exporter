# Configuration Guide

## Configs and Last Successful Folders

Inside the root folder of this script you'll find two folders:

* `configs`
* `last_successful`

### Configs
The `configs` folder holds each of your config files. The default is `config.yaml` but you can duplicate this file and create more if you need to. For most users, you won't need to but if you are pulling data for multiple users or organisations, you can create seperate config files in this folder and reference them when running the script. For example, you may make a copy of `config.yaml` and all it `new_org.yaml`. In which case, you'd run the script like this:

`python exporter.py --format csv --config new_org.yaml`

!!! Warning
    When you duplicate a config file, be sure to change the `config_name` option in the file. If you don't, your data could end up mixed together. 

### Last Successful
The `last_successful` folder contains text files, each appended with the name set in `config_name`. You'll most likely only have one file, and if you used the default config file, it'll be called `last_successful-iauditor`.

Inside this file is a single timestamp written in UTC format. It'll look like this: 

`2020-02-15T15:59:08.527Z`

The timestamp breaks down as:

* Year
* Month
* Day
* The letter 'T'
* Hours
* Minutes
* Seconds
* Microseconds (Must be three digits - if in doubt, set it to three zeros)

You can edit this timestamp to start the exporter from a given time. Let's say you only wanted to download inspections from 2020, you'd edit it like this:

`2020-01-01T00:00:00.000Z`

If you want to download everything from the beginning of your account, you can just delete the appropriate `last_successful` file entirely and the script will recreate it automatically, starting from the year 2000. 

## Editing the Config File

### Key Options for all formats

|  Setting | Optional? | Accepted Parameters |Description  |
|---|---| --- | --- |
| token | No | N/A | Your API key, generated from iAuditor. [Click here for guidance](https://support.safetyculture.com/integrations/how-to-get-an-api-token/)
| config_name | No | N/A | You can set the name of your configuration here. Very useful if you're managing multiple configurations as it'll be used to name files and organise folders. Do not use any spaces in this name. 
| export_completed | Yes | `true` `false` `both` | By default, we only export completed inspections from iAuditor. Set this to `true` to _only_ receive _completed_ audits, `false` to _only_ receive _incomplete_ inspections or `both` to export everything regardless of status. In the dataset, anything without a completed date is considered incomplete.  
| export_archived | Yes | `true` `false` `both` | By default, we do not export inspections . Set this to `true` to _only_ receive archived inspections, `false` to only receive _archived_ inspections or `both` to export everything regardless of status. In the dataset, the column `Archived` will be either `true` or `false` depending on the inspections current status.
| template_ids | Yes | See the tip below | Here you can specify the templates from which you'd like to download your data. Leave this option blank to export all available information. See the tip below for additional guidance.

!!! Tip
    * When setting template IDs you wish to export, you need to format the templates into a list like this: `template_123,template_456,template_789`. 
        
    * If you want just one template, just write it on its own like this: `template_123`.
           
    * If you have a large list of template IDs, you can save them into a `.txt` file. Place this in the same directory as `exporter.py` and enter the filename in this option (e.g. `templates.txt`). There is a limit on the number of templates you can place in a text file. If you receive errors you'll need to reduce it down or export everything and filter it afterwards.

### Options only for CSV

|  Setting | Optional? | Accepted Parameters |Description  |
|---|---| --- | --- |
| use_real_template_name | Yes | `true` `false` `single_file` | When exporting CSV files, we will export the files using template IDs. Setting this to `true` will override this and use the template name. You can also set `single_file` and the export will go to a single CSV file rather than splitting it across templates. This options has some caveats, please see the warning below.

!!! Warning for use_real_template_name
    We only recommend using the real template name if you are doing a one-off export. If a template is renamed, the script will create a new file rather than appending to the existing one. You also need to ensure that no two templates have the same name as the script would have no way to differentiate between the two. If in doubt, keep this as either `false` or `single_file`

### Options only for SQL

|  Setting | Optional? | Accepted Parameters |Description  |
|---|---| --- | --- |
| merge_rows/action_merge_rows | Yes | `true` `false` | This setting, when set to `true` will update existing rows in the database when an inspection is updated after being logged. There are important caveats to this option, please review the tips and warnings below.

!!! Warning
    There are some important caveats to enabling this option (namely it can make your exports slower, the dataset larger and cause increased database requests.) You should review [the model](../../understanding-the-data/the-model) documentation to fully understand how this works before enabling it. 

!!! Tip
    * This setting can only be set _before_ the database table is created. If you need to change this setting at a later date, you will need to drop the table first and allow it to be recreated. 
    * When enabled, `export_inactive_items` is forced to be `true` (Explained here: [the model](../../understanding-the-data/the-model))
    

### Other SQL Options

You must fill out all these configuration options to use the SQL export.

|  Setting | Description  |
|---|--- |
| sql_table |The name of the table in which you want to store your iAuditor information. Best practice is to make sure it doesn't exist, as the script will create it for you. If you want to build it yourself, check [the model](../..w/understanding-the-data/the-model).
| database_type |  For SQL use 'mssql+pyodbc_mssql' (More should work, however they're currently untested. please refer to the SQLAlchemy documentation) 
| database_user | The username to login to your database 
| database_pwd |  Your database password
| database_server | Server where your database is located
| database_port |  The port your database is listening on (For SQL, this is usually 1433)
| database_name |  The name of the database you'll be connecting to. You must also define the driver to use if you need to - for SQL you'll likely want to use `MyDatebase?driver=ODBC Driver 17 for SQL Server` - replacing `MyDatabase` with the name of your database. 


### Other Options

|  Setting | Optional? | Description  |
|---|---| --- |
export_path  | Yes | absolute or relative path to the directory where to save exported data to (this applies to everything except `SQL` exports)  |
| filename  |Yes |  an audit item ID whose response is going to be used to name the files of exported audit reports. Can only be an item with a response type of `text` from the header section of the audit such as Audit Title, Document No., Client / Site, Prepared By, Personnel, or any custom header item which has a 'text' type response (Only applies to PDF/Word Exports) |
| preferences  | Yes| to apply a preference transformation to particular templates, give here a list of preference ids. 
| sync_delay_in_seconds |Yes | time in seconds to wait after completing one export run, before running again
| export_inactive_items | Yes| This setting only applies when exporting to CSV. Valid values are true (export all items) or false (do not export inactive items). Items that are nested under [Smart Field](https://support.safetyculture.com/templates/smart-fields/) will be 'inactive' if the smart field condition is not satisfied for these items. This option is forced to `true` if you're using SQL and enable either of the `merge_rows` options.
| media_sync_offset_in_seconds | Yes | time in seconds since an audit has been modified before it will by synced
