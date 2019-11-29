#!/bin/bash

latestVersion=$(curl --silent "https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
echo ${latestVersion} > /home/pi/static/latestVersion
