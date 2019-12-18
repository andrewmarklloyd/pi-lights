#!/bin/bash

build() {
  GOOS=linux GOARCH=arm GOARM=5 go build -o lights main.go
}

if [[ ${1} == 'build' ]]; then
  build
fi
