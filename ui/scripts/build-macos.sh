#!/bin/bash

export PATH=${PATH}:`go env GOPATH`/bin
echo "building on AMD64"
wails build -platform darwin/amd64 -clean

echo "Zipping Package"
ditto -c -k --keepParent ./build/bin/safetyculture-exporter.app ./exporter-darwin-amd64.zip
echo "Cleaning up"
rm -rf build/bin/SafetyCulture Exporter.app

echo "building on ARM64"
wails build -platform darwin/arm64 -clean
echo "Zipping Package"
ditto -c -k --keepParent ./build/bin/safetyculture-exporter.app ./exporter-darwin-arm64.zip
echo "Cleaning up"
rm -rf ./build/bin/SafetyCulture Exporter.app
