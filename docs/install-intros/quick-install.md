# Quick Install

* Install Python 3.6+
* If you've used the SafetyCulture Python SDK before, you need to uninstall it (if you're unsure, run the command anyway): `pip uninstall safetyculture-sdk-python`
* Use Git to clone the repository and move to it in terminal/command prompt:
    * ` git clone https://github.com/SafetyCulture/iauditor-exporter.git`
    * `cd iauditor-exporter`
* Run `pip install -r requirements.txt`
* If you need SQL support, you'll need the driver installed. Install the relevant driver from [here](https://docs.microsoft.com/en-us/sql/connect/odbc/download-odbc-driver-for-sql-server?view=sql-server-ver15)
* Open `configs/config.yaml.sample`, rename it to `config.yaml` and edit as per the guidance [here](../../script-setup/config/)
* Run `python exporter.py --format sql` (Additional options available, see [here](../../script-setup/config/))

!!! Tip
    If you're on a Mac and installed Python using Brew, it's possible your Python install may be `python3` instead of just `python` and similarly `pip3` instead of `pip`