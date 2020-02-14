## Initial Steps

* If you've used the SafetyCulture Python SDK before, you need to uninstall it (if you're unsure, run the command anyway.):
    * pip uninstall safetyculture-sdk-python
* Clone or download the repository
* Navigate to `safetyculture-sdk-python` and run `pip install .`
* Navigate to `tools/exporter`
    * `cd safetyculture-sdk-python/tools/exporter`
* Run `pip install -r requirements.txt`
* Open `configs/config.yaml` and edit as needed
* Run `python exporter.py --format sql` (Or other format support by the exporter)