package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"

	"gopkg.in/yaml.v2"

	// "os/exec"

	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio"
)

type config struct {
	Server struct {
		Role       string `yaml:"role"`
		Pin        int    `yaml:"pin"`
		FollowerIP string `yaml:"followerIP"`
	} `yaml:"server"`
}

var pin rpio.Pin
var pinNumber int
var testmode bool
var cfg config

func main() {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println("unable to open config.yml", err)
		os.Exit(1)
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("unable to decode config file", err)
		os.Exit(1)
	}
	pinNumber = cfg.Server.Pin
	pin = rpio.Pin(pinNumber)

	err = rpio.Open()
	if err != nil {
		fmt.Println("unable to open gpio", err.Error())
		fmt.Println("running in test mode")
		testmode = true
	} else {
		fmt.Println("creating channel")
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanup(pin, pinNumber)
			os.Exit(1)
		}()

		pin.Output()

	}
	fmt.Println("Setting up http handlers")
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/switch", switchHandler)
	http.HandleFunc("/pin", pinHandler)
	http.HandleFunc("/system", systemHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func switchHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	op := req.FormValue("op")
	if testmode {
		fmt.Println(op)
	} else {
		if op == "on" {
			fmt.Println("ON")
			pin.Write(rpio.High)
		} else if op == "off" {
			fmt.Println("OFF")
			pin.Write(rpio.Low)
		}

		if cfg.Server.Role == "leader" {
			httpReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:8080/pin", cfg.Server.FollowerIP), nil)
			httpReq.Header.Set("op", op)
			var client = http.Client{Timeout: time.Second * 10}
			_, err := client.Do(httpReq)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
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

func systemHandler(w http.ResponseWriter, req *http.Request) {
	op := req.FormValue("op")
	var command string = ""
	if op == "shutdown" {
		command = "shutdown"
		fmt.Fprintf(w, "shutting down")
	} else if op == "reboot" {
		command = "reboot"
		fmt.Fprintf(w, "rebooting")
	} else if op == "update" {
		command = "./update.sh"
		fmt.Fprintf(w, "updating software")
	} else {
		fmt.Fprintf(w, "command not recognized")
	}
	fmt.Printf("Running command: %s\n", command)
	if command != "" && !testmode {
		if err := exec.Command(command).Run(); err != nil {
			fmt.Println("Failed to initiate command:", err)
		}
	}
}

func cleanup(pin rpio.Pin, pinNumber int) {
	fmt.Println("Cleaning up pin", pinNumber)
	pin.Write(rpio.Low)
	rpio.Close()
}
