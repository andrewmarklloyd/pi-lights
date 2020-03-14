# Pi Lights

[![Build Status](https://travis-ci.org/andrewmarklloyd/pi-lights.svg?branch=master)](https://travis-ci.org/andrewmarklloyd/pi-lights)

Small web server running on a Raspberry Pi that serves a web UI to control and schedule a connected relay.


### One Line Install
To install on a Raspberry Pi with a single line command, ssh to the Pi and run the following:
```
bash <(curl -s -H 'Cache-Control: no-cache' https://raw.githubusercontent.com/andrewmarklloyd/pi-lights/master/install/install.sh)
```

### Developing Locally
Requires Go 1.13.1 to build the project.
```
# run the program
go run main.go

# build an executable
go build -o lights main.go
```
