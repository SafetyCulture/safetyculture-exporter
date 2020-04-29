# Overview
The iAuditor Exporter tool is the primary way to bulk export iAuditor information for use in BI tools such as PowerBI. The tool is coded in the Python programming language and can be ran simply and easily on any computer with Python installed.

This documentation aims to make it easy to not only set up the tool, but to also understand and make use of the data it gives out.

## What's new?

If you've used this tool before, you're likely wondering what's new. Here's the changelog:

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