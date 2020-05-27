# Overview
The iAuditor Exporter tool is the primary way to bulk export iAuditor information for use in BI tools such as PowerBI. The tool is coded in the Python programming language and can be ran simply and easily on any computer with Python installed.

This documentation aims to make it easy to not only set up the tool, but to also understand and make use of the data it gives out.

## What's new?

### V2 - May 2020
* MySQL Support and improved database model. Should mean this will work with other databases, too. 
* Better error handling. Many common errors will not return useful hints or links to relevant documentation.
* Proxy support - if you use a proxy, you can specify them in your config file. See config.yaml.sample for where to add this option to your config file.
* SSL Certs - Initial support for custom SSL certificates. This has received limited testing, please let us know how you get on if you use this.
* If you export in multiple formats, everything but CSVs will be organised into relevant sub-folders
* The code has been split into multiple smaller modules to ease development
* Custom config files and last_successful files now work more reliably. 
* Fixed a bug where having empty config values caused the script to fail

### V1 - February 2020

* SQL Support - You can now export both inspection and action data directly into a SQL database
* Config-level options for exporting archived and incomplete inspections
* Config-level options to only export particular templates so you don't have to export everything if you don't want to
* If you're exporting in SQL, you have the option to merge the rows so if an inspection or action is updated, the database is updated rather than appended 
* Export in CSV to a single file rather than to multiple files
* Export CSVs with the template name rather than the template ID 
* An additional column to signify the archived status of an inspection
* An additional column for the parent ID of a question within a repeating section making it easier to group together questions within the same repeating section
* A new SortingIndex column to order questions should they fall out of order.  


If you feel any part of this documentation is lacking or missing, please reach out to support@safetyculture.com and let us know what you'd like to see.