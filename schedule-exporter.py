import sys
import subprocess

configs = ['config']

def start_export(config):
    # config = '--config {}.yaml'.format(config)
    subprocess.run('python exporter.py --format csv --config {}.yaml'.format(config), shell=True)

for config in configs:
    print('Starting {}'.format(config))
    start_export(config)