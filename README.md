# Pi Lights

[![Build Status](https://travis-ci.org/andrewmarklloyd/pi-lights.svg?branch=master)](https://travis-ci.org/andrewmarklloyd/pi-lights)

Small web server running on a Raspberry Pi that allows HTTP GET requests to trigger a 5v relay connected to a string of Christmas lights.


### One Line Install
To install on a Raspberry Pi with a single line command, run the following:
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
