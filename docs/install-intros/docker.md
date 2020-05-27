# Docker 

## Introduction to Docker
Docker is a useful way to run software within what's known as 'containers'. By running in a container, you do not need to install any requirements such as Python or its associated package. All the requirements are already 'contained' with the container, meaning it will run exactly the same on different systems. 

!!! Info
    The Docker container contains [Microsoft's ODBC for SQL Server](https://docs.microsoft.com/en-us/sql/connect/odbc/microsoft-odbc-driver-for-sql-server?view=sql-server-ver15) - by using the container, you're agreeing to Microsoft's EULA for the driver. 

## Why use Docker?

### Advantages
* No additional requirements to install - as long as you have Docker installed, you shouldn't need anything else. 
* If the script fails, Docker can automatically restart it without any user intervention 
* It's easier to run the script as a background service when using Docker
* If you don't have a database server, you could store your data inside a Docker database

### Disadvantages
* Docker is more resource intensive than the Python script run on its own

## Deploying the Exporter Tool with Docker

The Exporter Tool contains a lot of configuration options so it's highly recommended to use Docker Compose to deploy the script. There is an example Docker Compose file in the repository already, it's called `docker-compose.yml` and looks like this:

```
version: '3.7'

services:
  iauditor-exporter:
    image: eddabrahamsen/iauditor-exporter:latest
    container_name: iauditor-exporter
    restart: unless-stopped
    environment:
      - PGID=1000
      - PUID=1000
      - 'format=sql actions-sql'
      - CONFIG_NAME=iauditor
      - API_TOKEN=
      - SYNC_DELAY_IN_SECONDS=900
      - MEDIA_SYNC_OFFSET_IN_SECONDS=900
      - TEMPLATE_IDS=
      - SQL_TABLE=iauditor_data
      - DB_TYPE=mssql+pyodbc_mssql
      - DB_USER=
      - DB_PWD=
      - DB_SERVER=
      - DB_PORT=1433
      - DB_NAME=iAuditor?driver=ODBC Driver 17 for SQL Server
      - DB_SCHEMA=dbo
      - USE_REAL_TEMPLATE_NAME=false
      - EXPORT_ARCHIVED=false
      - EXPORT_COMPLETED=both
      - MERGE_ROWS=false
      - ACTIONS_MERGE_ROWS=false
      - ALLOW_TABLE_CREATION=false
    volumes:
      - type: bind
        source: ./last_successful
        target: /app/last_successful
      - type: bind
        source: ./exports
        target: /app/exports
  portainer:
    container_name: portainer
    image: portainer/portainer
    restart: unless-stopped
    ports:
      - "9000:9000"
    command: -H unix:///var/run/docker.sock
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - iauditor_portainer_data:/data
volumes:
  iauditor_portainer_data:

```

The first thing to do is fill in the configuration options within `environment`. These ones:
```
    environment:
      - PGID=1000
      - PUID=1000
      - 'format=sql actions-sql'
      - CONFIG_NAME=iauditor
      - API_TOKEN=
      - SYNC_DELAY_IN_SECONDS=900
      - MEDIA_SYNC_OFFSET_IN_SECONDS=900
      - TEMPLATE_IDS=
      - SQL_TABLE=iauditor_data
      - DB_TYPE=mssql+pyodbc_mssql
      - DB_USER=
      - DB_PWD=
      - DB_SERVER=
      - DB_PORT=1433
      - DB_NAME=iAuditor?driver=ODBC Driver 17 for SQL Server
      - DB_SCHEMA=dbo
      - USE_REAL_TEMPLATE_NAME=false
      - EXPORT_ARCHIVED=false
      - EXPORT_COMPLETED=both
      - MERGE_ROWS=false
      - ACTIONS_MERGE_ROWS=false
      - ALLOW_TABLE_CREATION=false

```

Most of these settings are explained on the [config options page](../../script-setup/config/) however there are three Docker specific options:

#### PGID and PUID
These are Docker specific options for the UID of the user/group running the container. If in doubt, leave it as 1000. If you know you need to change this, here's the place to do it. 

#### 'format='
This option dictates what data the tool will export. In this example, we'll be exporting in `sql` (inspection data) and also `actions-sql` (actions in SQL). You can set this to any of the config options discussed [here](../../script-running/export-options/). The main thing to be mindful of is ensuring you keep the surrounding apostrophes. 

#### ALLOW_TABLE_CREATION
If you're using SQL, when connecting to your server, we check if your specified table (`SQL_TABLE`) exists. If it doesn't, we'll attempt to create it using our [model](../../understanding-the-data/the-model). However, it's important that we don't make changes to your SQL server without express permission which is why this option defaults to `false`. If you want us to create the table for you, you must change this option to `true`. 

The completed section may look like this:

```
    environment:
      - PGID=1000
      - PUID=1000
      - 'format=sql actions-sql'
      - CONFIG_NAME=iauditor
      - API_TOKEN=b7f8f791920c1618ace0e24b4d52ce260473dad870e7bd56b869f8d2f26e554d
      - SYNC_DELAY_IN_SECONDS=900
      - MEDIA_SYNC_OFFSET_IN_SECONDS=900
      - TEMPLATE_IDS=template_123,template_456
      - SQL_TABLE=iauditor_data
      - DB_TYPE=mssql+pyodbc_mssql
      - DB_USER=SA
      - DB_PWD=pa55w0rd
      - DB_SERVER=localhost
      - DB_PORT=1433
      - DB_NAME=iAuditor?driver=ODBC Driver 17 for SQL Server
      - DB_SCHEMA=dbo
      - USE_REAL_TEMPLATE_NAME=false
      - EXPORT_ARCHIVED=false
      - EXPORT_COMPLETED=both
      - MERGE_ROWS=false
      - ACTIONS_MERGE_ROWS=false
      - ALLOW_TABLE_CREATION=false

```

### Running with Compose
* Ensure you're in the correct directory first: `cd iauditor-exporter` - or wherever you downloaded the script to. 
* Run `docker-compose up iauditor-exporter`
* If your config file has been filled out correctly, the export should begin. If you get any errors, check your config file. 
* Assuming the export begins, we can look at running this more permanently. 

## Daemon Mode
By running the container in daemon mode, we can create a service which runs in the background. It doesn't require us to keep a terminal window open, and we can manage its progress through the web browser using a tool called Portainer. 

If your export from before is still running from before, press `CTRL+C` on your keyboard to stop it. 

Now, run: `docker-compose up -d`

This will start the exporter, but also start up a very useful tool called Portainer. If you do not want to use Portainer, you can simply run `docker-compose up -d iauditor-exporter` or remove Portainer from your compose file. 


### Portainer
Portainer is a useful tool for monitoring the activity of Docker Containers in your web browser. After the above commands have finished, you can open your web browser and visit [http://localhost:9000](http://localhost:9000)
* Portainer will prompt you to create a login - you can set this to whatever you like. 
* Click "Containers"
* You should see `iauditor-exporter` in the list of running containers 
* From here, you can easily stop, start, restart and pause the container amongst other things.    