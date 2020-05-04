# Export Options

Although the primary use for this tool is to export inspection data into either CSV or SQL, there are other formats supported. 

You can run any of these options either individually or together by adding them with a space when running the tool:

`python exporter.py --format sql actions-sql media`

In the above example, you'd export all your inspection data and actions to SQL, as well as locally downloading all associated media. 
## Formats 

### CSV

`python exporter.py --format csv`

Exports all available inspection data into CSV format. See [here](../../understanding-the-data/datastructure/) for detailed discussion of the output. 

### SQL

`python exporter.py --format sql`

Exports all available inspection data into the configured SQL database. 

* See [here](../../understanding-the-data/datastructure/) for detailed discussion of the output
* See [here](../../understanding-the-data/the-model/) for detailed discussion of how we structure the SQL table

### Actions
`python exporter.py --format actions`

Exports all available Actions into a CSV table. 

### Actions SQL
`python exporter.py --format actions-sql`

Exports all available Actions into a SQL table. 

### PDF/DOCX

`python exporter.py --format pdf` or `python exporter.py --format docx`

Or combine them and have both: `python exporter.py --format pdf docx`


Exports all available inspections in PDF and/or DOCX (Word) format. 

### Media
Exports all the media attached to available inspections and stores them in a folder per inspection. Files are named with their unique media ID + the file extension of the image. More information on this can be found in the tip [here](../../understanding-the-data/datastructure/#mediahypertextreference). 

### Web Report Links
`python exporter.py --format web-report-link`

Exports a CSV file of AuditIDs along with their public web report links. This is really useful when working with tools like PowerBI, as you can use it as a reference table for web report links without having to store them multiple times in the main dataset. SQL support for this table is coming soon. 

The CSV file includes five columns: Template ID, Template Name, Audit ID, Audit Name, and Web Report Link.

!!! Warning
    By running the tool with this option, we'll generate public links for every inspection available. Although these links will not be tracked by search engines, they are still completely public by design. If you do not want any of your inspections to have generated public URLs, do not use this option. 

### JSON

Exports all available inspections in JSON format. The output here is no different to downloading inspections directly from the API. 


## Additional Parameters

### Loop
`python exporter.py --format sql --loop`

By default, if there are more than 1000 inspections to export or once all inspections have exported, the tool will exit. By adding `--loop` when running the tool, the script will wait the number of seconds defined by `sync_delay_in_seconds` in the config file (default is 900 seconds/15 minutes) and continue.

### Config
`python exporter.py --format sql --config new_config.yaml`

You may need to maintain multiple configurations, in which case you can define the name of the configuration file you wish to use. By default we look in the `configs` folder for a file called `config.yaml`. If you want to use a different file, just drop it into the `configs` folder and reference it here. 

### 