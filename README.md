# iAuditor Exporter tool

> ## **PLEASE NOTE: This iAuditor Exporter tool version is no longer maintained.**
> 
> We encourage you to download and install the [latest version of the iAuditor Exporter](https://github.com/SafetyCulture/iauditor-exporter).

---

The iAuditor Exporter tool is a data export script that’s available to all our Premium customers. It’s written in the Python programming language and can be easily installed and run on any computer with Python installed. Although it’s primarily used to bulk export inspections to CSV and SQL formats for business intelligence tools, it’s also capable of exporting data to formats that you can find on the web app and the mobile app, such as [PDF](https://help.safetyculture.com/en_us/1064141989-HJTwSeJIw), [Word](https://help.safetyculture.com/en_us/1064141989-HJTwSeJIw), and [web report links](https://help.safetyculture.com/en_us/1063814672-BkzsulyUv).

This README shows you how to install and run the iAuditor Exporter tool to bulk export iAuditor data. As the iAuditor Exporter tool utilizes our software development kit (SDK) to interact with our API and export data, you can also use our SDK in either [Python](https://github.com/SafetyCulture/safetyculture-sdk-python) or [JavaScript](https://github.com/SafetyCulture/safetyculture-js) to build your own custom integrations

## Before you begin

Please note that you must be on our [Premium subscription](https://safetyculture.com/pricing/) to install and run the iAuditor Exporter tool.

As the script runs on command lines, it’s best to have some basic knowledge of running command lines before installing the iAuditor Exporter tool.

Please follow the instructions in this README carefully and take note when the steps that differ between macOS (Apple) and Windows computers. If you run into any errors or have any questions regarding the instructions, please [contact our customer support team](https://safetyculture.com/contact-us/) for assistance.

## 1\. Install Python

### macOS

1.  Open the “Terminal” app on your computer.
2.  [Install Homebrew](https://brew.sh/).
3.  Run the following command line:  
    `brew install python3`  
    If your computer is running macOS 10.15 or above, you may already have Python 3 installed. Learn how to [check the Python version on macOS](https://installpython3.com/mac/).
4.  Run the following command line:  
    `pip install pipx`  
    If the above command line returns an error, run the following command line:  
    `pip3 install pipx`  
    If the above command line returns an error, restart Terminal and try again.
5.  Proceed to install the iAuditor Exporter tool.

### Windows

1.  [Download and install the latest version of Python for Windows](https://www.python.org/downloads/windows/). We recommend that you download the Windows x86-64 executable installer.
2.  During installation, make sure to check the “Add to PATH” box. If you forgot to select the checkbox, re-run the installer and try again.
3.  Open the “PowerShell” app on your computer.
4.  Run the following command line:  
    `pip install pipx`
5.  Proceed to install the iAuditor Exporter tool.

## 2\. Install the iAuditor Exporter tool

1.  In your command line window, navigate to the location where you want to store the iAuditor Exporter tool. Learn how to change directory or folder using command lines on [macOS](https://github.com/0nn0/terminal-mac-cheatsheet#core-commands) and [Windows](https://docs.microsoft.com/windows-server/administration/windows-commands/cd).
2.  Run the following command line:  
    `pipx install iauditor_exporter`
3.  Once the installation is complete, run the following command line:  
    `ia_exporter --setup`
4.  In the command line window, user arrow keys to navigate and the enter or return key to select an option for configuration. For a basic configuration, modify the following options:
    *   **Change time to search from**: Determines the date and time in Coordinated Universal Time (UTC) to start exports. Leave blank if you want to export all-time data from your iAuditor account.
    *   **token**: Determines the iAuditor account to use for exports.
5.  Select “Exit and save” once the configuration is complete. You can also manually configure the options in the “configs” folder’s “config.yaml” file.
6.  Proceed to run the iAuditor Exporter tool.

## 3\. Run the iAuditor Exporter tool

Each time you run an export, remember to navigate to the directory where you installed the iAuditor Exporter tool. Learn how to change directory or folder using command lines on [macOS](https://github.com/0nn0/terminal-mac-cheatsheet#core-commands) and [Windows](https://docs.microsoft.com/windows-server/administration/windows-commands/cd).

1.  Run the following command line to bulk export inspections to CSV:  
    `ia_exporter --format csv`
2.  Keep in mind that the iAuditor Exporter tool can only process up to 1,000 inspections each run. You can set it to run again automatically by adding “--loop” to the command line:  
    `ia_exporter --format csv --loop`  
    This sets the iAuditor Exporter tool to wait for a period of time after each run before starting again, which is defined in the “config.yaml” file’s “sync\_delay\_in\_seconds”.
3.  By default, the iAuditor Exporter tool downloads 100 inspections at a time. Depending on the size of the export you are running, you have the option to increase or decrease the number of inspection downloads by adding “--chunks” to the command line and specifying the size.
    *   Run the following command line with a number greater than 100 to increase the inspection download size:  
        `ia_exporter --format csv --chunks 500`
    *   Run the following command line with a number less than 100 to decrease the inspection download size:  
        `ia_exporter --format csv --chunks 50`

## Export options

Although the iAuditor Exporter tool is primarily used to bulk export inspections to CSV and SQL formats, it can also export data to the following formats.

*   **Actions**: Exports all available actions into a CSV spreadsheet.  
    `ia_exporter --format actions`
*   **Actions SQL**: Exports all available actions into the specified SQL database. Make sure to follow the instructions in this README to add database support.  
    `ia_exporter --format actions-sql`
*   **CSV**: Exports all available inspection data into a CSV spreadsheet.  
    `ia_exporter --format csv`
*   **JSON**: Exports all available inspections to individual JSON files.  
    `ia_exporter --format json`
*   **Media**: Exports all available attached media files into individual inspection folders.  
    `ia_exporter --format media`
*   **PDF**: Exports all available inspections to individual PDF files.  
    `ia_exporter --format pdf`
*   **SQL**: Exports all available inspection data into the specified SQL database. Make sure to follow the instructions in this README to add database support.  
    `ia_exporter --format sql`
*   **Web report links**: Generates and lists [web report links](https://help.safetyculture.com/1063814672-BkzsulyUv) for all available inspections in a CSV spreadsheet.  
    `ia_exporter --format web-report-link`
*   **Word**: Exports all available inspections to individual Word files.  
    `ia_exporter --format docx`

## Database support

The iAuditor Exporter tool includes support for SQL, PostgreSQL, and MySQL databases. Depending on the system and database you use, you may need to satisfy the following dependencies before running the iAuditor Exporter tool.

*   **Windows**: No actions required.
*   **macOS**: The following dependencies are can be [installed via Brew](https://brew.sh/).
    *   `unixodbc`
    *   MySQL: `mysql`
    *   PostgreSQL: `postgres`
*   **Linux**: You may need to adapt the following dependencies, as they assume the Debian or Ubuntu system.
    *   `unixodbc-dev`
    *   MySQL: `libmysqlclient-dev`
    *   MariaDB: `libmariadbclient-dev`
    *   PostgreSQL: `libpq-dev`

Once you’ve installed the dependencies for your system and database, run the following command lines to add support.

*   SQL Server  
    `pipx install "iauditor_exporter[sql]"`
*   PostgreSQL  
    `pipx install "iauditor_exporter[postgres]"`
*   MySQL  
    `pipx install "iauditor_exporter[mysql]"`
*   All three databases. Keep in mind that you must meet the additional requirements for this option, so only proceed if you require support for all three databases.  
    `pipx install "iauditor_exporter[all_db]"`
