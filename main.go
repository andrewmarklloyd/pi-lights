package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stianeikeland/go-rpio"
)

type config struct {
	Server struct {
		Role       string `yaml:"role"`
		Pin        int    `yaml:"pin"`
		FollowerIP string `yaml:"followerIP"`
	} `yaml:"server"`
}

type AppInfo struct {
	TagName string `json:"tag_name"`
}

type HomePageData struct {
	Version       string
	LatestVersion string
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

	version, err := ioutil.ReadFile("static/version")
	if err != nil {
		fmt.Println("unable to open verison", err)
		os.Exit(1)
	}

	c := cron.New()
	c.AddFunc("@every 45m", checkForUpdates)
	c.Start()

	fmt.Println("Setting up http handlers")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		d, _ := ioutil.ReadFile(string(path))
		if strings.HasSuffix(path, ".css") {
			w.Header().Add("Content Type", "text/css")
			w.Write(d)
		} else if path == "" {
			latestVersion, err := ioutil.ReadFile("static/latestVersion")
			if err != nil || len(latestVersion) == 0 {
				latestVersion = version
			}
			tmpl := template.Must(template.ParseFiles("./static/index.html"))
			data := HomePageData{
				Version:       string(version),
				LatestVersion: string(latestVersion),
			}
			tmpl.Execute(w, data)
		}
	})
	http.HandleFunc("/switch", switchHandler)
	http.HandleFunc("/pin", pinHandler)
	http.HandleFunc("/system", systemHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func checkForUpdates() {
	fmt.Println("Checking for updates")
	resp, _ := http.Get("https://api.github.com/repos/andrewmarklloyd/pi-lights/releases/latest")
	var info AppInfo
	err := json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Writing latestVersion to file")
		versionInfo := []byte(info.TagName)
		err = ioutil.WriteFile("./static/latestVersion", versionInfo, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}
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
		command = "/home/pi/install/update.sh"
		fmt.Fprintf(w, "updating software")
	} else if op == "check-updates" {
		checkForUpdates()
		fmt.Fprintf(w, "checking for updates")
	} else {
		fmt.Fprintf(w, "command not recognized")
	}
	fmt.Printf("Running command: %s\n", command)
	if command != "" && !testmode {
		cmd := exec.Command(command)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Start()
		if err != nil {
			fmt.Println("Failed to initiate command:", err)
			os.Exit(1)
		}
		fmt.Printf("Command: %q\n", out.String())
	}
}

func cleanup(pin rpio.Pin, pinNumber int) {
	fmt.Println("Cleaning up pin", pinNumber)
	pin.Write(rpio.Low)
	rpio.Close()
}
