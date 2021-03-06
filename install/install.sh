#!/bin/bash

sudo apt-get update
sudo apt-get install jq -y

archive_path="/tmp/pi-lights"
install_dir="/home/pi"
mkdir -p ${archive_path}

latestVersion=$(curl --silent "https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -sL https://github.com/andrewmarklloyd/pi-lights/archive/${latestVersion}.tar.gz | tar xvfz - -C "${archive_path}" --strip 1 > /dev/null

binaryUrl=$(curl -s https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest | jq -r '.assets[] | select(.name == "lights") | .browser_download_url')
curl -sL $binaryUrl -o ${archive_path}/lights
chmod +x ${archive_path}/lights
rm -f ${install_dir}/install/*
rm -f ${install_dir}/static/*

mkdir -p ${install_dir}/install/
mkdir -p ${install_dir}/static/
cp ${archive_path}/install/* ${install_dir}/install/
cp ${archive_path}/static/* ${install_dir}/static/
mv ${archive_path}/default.config.yml ${install_dir}/config.yml

echo -n ${latestVersion} > ${install_dir}/static/version
echo -n ${latestVersion} > ${install_dir}/static/latestVersion
mv ${archive_path}/lights ${install_dir}/

echo "Enter token to be used for API authentication, then press enter"
read -s token
sed -i "s/{{.token}}/${token}/g" ${archive_path}/install/lights.service
sudo mv ${archive_path}/install/lights.service /etc/systemd/system/
rm -rf ${archive_path}

sudo systemctl enable lights.service
sudo systemctl start lights.service

install_pagekite() {
  echo "Enter the kite name:"
  read KITE_NAME
  curl -O https://pagekite.net/pk/pagekite.py
  chmod +x pagekite.py
  sudo mv -f pagekite.py /usr/local/bin
  # TODO: configure kite here for first time
  sed "s/{{.kiteName}}/${KITE_NAME}/g" ${archive_path}/install/pagekite.service
  sudo mv ${archive_path}/install/pagekite.service /etc/systemd/system/
  sudo systemctl enable lights.service
  sudo systemctl start lights.service
}

echo "Would you like to enable pagekite service? [y/n]"
read answer
if [[ ${answer} == "y" ]]; then
  install_pagekite
fi
