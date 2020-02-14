# Welcome to MkDocs

For full documentation visit [mkdocs.org](https://mkdocs.org).

## Commands

* `mkdocs new [dir-name]` - Create a new project.
* `mkdocs serve` - Start the live-reloading docs server.
* `mkdocs build` - Build the documentation site.
* `mkdocs help` - Print this help message.

## Project layout

    mkdocs.yml    # The configuration file.
    docs/
        index.md  # The documentation homepage.
        ...       # Other markdown pages, images and other files.

## Installation
* If you've used the SafetyCulture Python SDK before, you need to uninstall it (if you're unsure, run the command anyway.):
    * pip uninstall safetyculture-sdk-python
* Clone or download the repository
* Navigate to `safetyculture-sdk-python` and run `pip install .`
* Navigate to `tools/exporter`
    * `cd safetyculture-sdk-python/tools/exporter`
* Run `pip install -r requirements.txt`
* Open `configs/config.yaml` and edit as needed
* Run `python exporter.py --format sql` (Or other format support by the exporter)



## Structure
 - Understanding the Data Set