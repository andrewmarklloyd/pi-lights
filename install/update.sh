#!/bin/bash

download_latest_release() {
  assetUrls=$(curl -s https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest | jq -r ".assets[] | .browser_download_url")

  for url in ${assetUrls}; do
    curl -sLO $url
  done
  chmod +x lights
}

sudo systemctl stop lights.service
download_latest_release
sudo systemctl start lights.service
