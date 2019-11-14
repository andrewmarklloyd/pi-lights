package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio"
)

var pin rpio.Pin
var pinNumber int
var hostname string

func main() {
	hostname, _ = os.Hostname()
	data, _ := ioutil.ReadFile("./config")
	s := strings.Trim(string(data), "\n")
	pinNumber, _ := strconv.Atoi(s)
	pin = rpio.Pin(pinNumber)

	err := rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}

	fmt.Println("creating channel")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(pin, pinNumber)
		os.Exit(1)
	}()

	pin.Output()
	fmt.Println("Setting up http handlers")
	http.HandleFunc("/switch", switchHandler)
	http.HandleFunc("/pin", pinHandler)
	http.ListenAndServe(":8080", nil)
}

func switchHandler(w http.ResponseWriter, req *http.Request) {
	pin.Toggle()
	pinStatus := pin.Read()
	var op string
	fmt.Println("pin", pinStatus)
	var status string
	if pinStatus == 0 {
		status = "<h1>OFF</h1>"
		op = "off"
	} else if pinStatus == 1 {
		status = "<h1>ON</h1>"
		op = "on"
	}

	if hostname != "zero" {
		httpReq, _ := http.NewRequest("GET", "http://192.168.0.115:8080/pin", nil)
		httpReq.Header.Set("op", op)
		var client = http.Client{Timeout: time.Second * 10}
		resp, err := client.Do(httpReq)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(resp.StatusCode)
		}
	}
	fmt.Fprintf(w, status)
}

func pinHandler(w http.ResponseWriter, req *http.Request) {
	op := req.Header.Get("op")
	var status string
	if op == "on" {
		pin.Write(rpio.High)
		status = "ON"
	} else if op == "off" {
		pin.Write(rpio.Low)
		status = "OFF"
	}
	fmt.Fprintf(w, status)
}

func cleanup(pin rpio.Pin, pinNumber int) {
	fmt.Println("Cleaning up pin", pinNumber)
	pin.Write(rpio.Low)
	rpio.Close()
}
