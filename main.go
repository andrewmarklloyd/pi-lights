package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/stianeikeland/go-rpio"
)

var pin rpio.Pin

func main() {
	pin = rpio.Pin(4)

	err := rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}

	fmt.Println("creating channel")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(pin)
		os.Exit(1)
	}()

	pin.Output()
	fmt.Println("Setting up http handlers")
	http.HandleFunc("/switch", switchHandler)
	http.ListenAndServe(":8080", nil)
}

func switchHandler(w http.ResponseWriter, req *http.Request) {
	pin.Toggle()
	pinStatus := pin.Read()
	fmt.Println("pin", pinStatus)

	if pinStatus == 0 {
		fmt.Fprintf(w, "<h1>OFF</h1>")
	} else if pinStatus == 1 {
		fmt.Fprintf(w, "<h1>ON</h1>")
	}
}

func cleanup(pin rpio.Pin) {
	fmt.Println("Cleaning up")
	pin.Write(rpio.Low)
	rpio.Close()
}
