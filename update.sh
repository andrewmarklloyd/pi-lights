#!/bin/bash

download_latest_release() {
  assetUrls=$(curl -s https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest | jq -r ".assets[] | .browser_download_url")

  for url in ${assetUrls}; do
    curl -sLO $url
  done
}

download_latest_release
