# Configs and Last Successful

Inside the root folder of this script you'll find two folders:

* `configs`
* `last_successful`

# Configs
The `configs` folder holds each of your config files. The default is `config.yaml` but you can duplicate this file and create more if you need to. For most users, you won't need to but if you are pulling data for multiple users or organisations, you can create seperate config files in this folder and reference them when running the script. For example, you may make a copy of `config.yaml` and all it `new_org.yaml`. In which case, you'd run the script like this:

`python exporter.py --format csv --config new_org.yaml`

!!! Warning
    When you duplicate a config file, be sure to change the `config_name` option in the file. If you don't, your data could end up mixed together. 

# Last Successful
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