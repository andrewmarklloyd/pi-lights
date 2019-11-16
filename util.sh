#!/bin/bash

build() {
  GOOS=linux GOARCH=arm GOARM=5 go build -o lights main.go
}

install() {
  scp -pr ./static user@host:the/target/directory
  scp lights ${username}@${host}:
  scp lights.service ${username}@${host}:
  ssh ${username}@${host} "sudo mv lights.service /etc/systemd/system/; sudo systemctl enable lights.service; sudo systemctl start lights.service"
}

deploy() {
  ssh ${username}@${host} "sudo systemctl stop lights.service"
  scp lights ${username}@${host}:
  scp -pr ./static ${username}@${host}:
  ssh ${username}@${host} "sudo systemctl restart lights.service"
}

if [[ ${1} == 'install' ]]; then
  build
  install
elif [[ ${1} == 'build' ]]; then
  build
elif [[ ${1} == 'deploy' ]]; then
  username=${2}
  host=${3}
  if [[ -z ${username} || -z ${host} ]]; then
    echo "Use username and host as arguments. Example:"
    echo "./util.sh deploy pi raspberrypi.local"
    exit 1
  fi
  build
  deploy
fi
