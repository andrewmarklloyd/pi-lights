#!/bin/bash

sudo systemctl stop lights.service

archive_path="/tmp/pi-lights"
mkdir -p ${archive_path}

latestVersion=$(curl --silent "https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sL https://github.com/andrewmarklloyd/pi-lights/archive/${latestVersion}.tar.gz | tar xvfz - -C "${archive_path}" --strip 1 > /dev/null

binaryUrl=$(curl -s https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest | jq -r '.assets[] | select(.name == "lights") | .browser_download_url')
curl -sL $binaryUrl -o ${archive_path}/lights
chmod +x ${archive_path}/lights
rm -f ./install/*
rm -f ./static/*
cp ${archive_path}/install/* install/
cp ${archive_path}/static/* static/

rm -rf ${archive_path}

sudo systemctl start lights.service
