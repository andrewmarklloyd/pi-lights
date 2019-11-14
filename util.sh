#!/bin/bash

build() {
  GOOS=linux GOARCH=arm GOARM=5 go build -o lights main.go
}

deploy() {
  scp lights ${username}@${host}:
  scp lights.service ${username}@${host}:
  ssh ${username}@${host} "sudo cp lights.service /etc/systemd/system/; rm lights.service; sudo systemctl enable lights.service; sudo systemctl enable lights.service; sudo systemctl start lights.service"
  curl http://${host}:8080/switch
}

if [[ ${1} == 'build' ]]; then
  build
elif [[ ${1} == 'deploy' ]]; then
  username=${2}
  host=${3}
  if [[ -z ${username} || -z ${host} ]]; then
    echo "Use username and host as arguments. Example:"
    echo "./util.sh deploy pi raspberrypi.local"
    exit 1
  fi
  deploy
fi
