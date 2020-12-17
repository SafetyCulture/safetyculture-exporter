import subprocess

configs = ["config"]


def start_export(selected_config):
    subprocess.run(
        "python exporter.py --format csv --config {}.yaml".format(selected_config),
        shell=True,
    )


for config in configs:
    print("Starting {}".format(config))
    start_export(config)
