## Editing the Config File

### Key Options

There are many options available, however only two are required for the script to run and export CSV files. 

|  Setting | Optional? | Description  |
|---|---| --- |
| token | No | Your API key, generated from iAuditor. [Click here for guidance](https://support.safetyculture.com/integrations/how-to-get-an-api-token/)
| config_name | No | You can set the name of your configuration here. Very useful if you're managing multiple as it'll be used to name files and organise folders. Do not use any spaces in this name. 

### SQL Specific Options

|  Setting | Description  |
|---|--- |
| sql_table |The name of the table in which you want to store your iAuditor information. Best practice is to make sure it doesn't exist, as the script will create it for you. 
| database_type |  The type of database you're using. for SQL use 'mssql+pyodbc_mssql', for Postgres it's 'postgresql' (More should work, please refer to the SQLAlchemy documentation) 
| database_user | The username to login to your database
| database_pwd |  Your database password
| database_server | Server where your database is located
| database_port |  The port your database is listening on
| database_name |  The name of the database you'll be connecting to. You can also define the driver to use if you need to - for SQL you'll likely want to add `?driver=ODBC Driver 17 for SQL Server`


### Other Options

|  Setting | Optional? | Description  |
|---|---| --- |
export_path  | Yes | absolute or relative path to the directory where to save exported data to  |
| filename  |Yes |  an audit item ID whose response is going to be used to name the files of exported audit reports. Can only be an item with a response type of `text` from the header section of the audit such as Audit Title, Document No., Client / Site, Prepared By, Personnel, or any custom header item which has a 'text' type response (doesn't apply when exporting as CSV) |
| use_real_template_name | Yes | If you set this to true, the script will append the name of the template to the exported CSV file. Keep in mind that if you use this option and change the name of a template, a new file will be generated next time the script runs. |
| preferences  | Yes| to apply a preference transformation to particular templates, give here a list of preference ids
| template_ids | Yes | Here you can specify the templates from which you'd like to download your data. You need to format the templates into a list like this: `template_123,template456,template,789` - If you want just one template, just write it on its own, like this: `template_123`
| sync_delay_in_seconds |Yes | time in seconds to wait after completing one export run, before running again
| export_inactive_items | Yes| This setting only applies when exporting to CSV. Valid values are true (export all items) or false (do not export inactive items). Items that are nested under [Smart Field](https://support.safetyculture.com/templates/smart-fields/) will be 'inactive' if the smart field condition is not satisfied for these items.
| media_sync_offset_in_seconds | Yes | time in seconds since an audit has been modified before it will by synced
