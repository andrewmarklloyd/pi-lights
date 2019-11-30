# Pi Lights

[![Build Status](https://travis-ci.org/andrewmarklloyd/pi-lights.svg?branch=master)](https://travis-ci.org/andrewmarklloyd/pi-lights)

Small web server running on a Raspberry Pi that allows HTTP GET requests to trigger a 5v relay connected to a string of Christmas lights.

### Build and Deploy
Requires Go 1.13.1 to build the project. The `util.sh` script will build and copy the binary to a remote Raspberry Pi.
```
# build an executable
./util.sh build

# copy executable and systemd config, start service
./util.sh pi raspberrypi.local
```
