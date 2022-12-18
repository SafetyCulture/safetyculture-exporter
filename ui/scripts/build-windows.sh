#!/bin/bash

wails build -platform windows/amd64 -clean

tar -czf exporter-windows-amd64.tar.gz ./build/bin/safetyculture-exporter.exe
