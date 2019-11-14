#!/bin/bash

username=${1}
host=${2}
if [[ -z ${username} || -z ${host} ]]; then
  echo "use username and host as arguments"
  exit 1
fi

build() {
  GOOS=linux GOARCH=arm GOARM=5 go build -o lights main.go
}

deploy() {
  scp lights ${username}@${host}:
  scp lights.service ${username}@${host}:
  ssh ${username}@${host} "sudo cp lights.service /etc/systemd/system/; rm lights.service; sudo systemctl enable lights.service; sudo systemctl enable lights.service; sudo systemctl start lights.service"
  curl http://${host}:8080/switch
}

ssh ${username}@${host} "sudo systemctl stop lights.service"
build
if [[ $? == 0 ]]; then
  deploy
fi
