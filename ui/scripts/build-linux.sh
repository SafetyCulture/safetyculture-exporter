#! /bin/bash

wails build -platform linux/amd64 -clean

tar -czf exporter-linux-amd64.tar.gz ./build/bin/safetyculture-exporter
