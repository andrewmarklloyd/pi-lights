package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"text/template"

	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/stianeikeland/go-rpio"
	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Server struct {
		Role       string `yaml:"role"`
		Pin        int    `yaml:"pin"`
		FollowerIP string `yaml:"followerIP"`
		Debug      bool   `yaml:"debug"`
	} `yaml:"server"`

	Schedule Schedule `yaml:"schedule"`
}

type Schedule struct {
	OnHour     string `yaml:"onHour"`
	OnMinutes  string `yaml:"onMinutes"`
	OffHour    string `yaml:"offHour"`
	OffMinutes string `yaml:"offMinutes"`
}

type AppInfo struct {
	TagName string `json:"tag_name"`
}

type HomePageData struct {
	Version       string
	LatestVersion string
	Debug         bool
	Schedule      Schedule
}

var pin rpio.Pin
var pinNumber int
var testmode bool
var cfg config
var cronLib *cron.Cron

func main() {
	cfg = readConfig()
	pinNumber = cfg.Server.Pin
	pin = rpio.Pin(pinNumber)

	err := rpio.Open()
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
		fmt.Println("unable to open version", err)
		os.Exit(1)
	}

	configureCron(cfg.Schedule)

	fmt.Println("Setting up http handlers")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]

		if path == "" {
			latestVersion, err := ioutil.ReadFile("static/latestVersion")
			if err != nil || len(latestVersion) == 0 {
				latestVersion = version
			}
			cfg = readConfig()
			tmpl := template.Must(template.ParseFiles("./static/index.html"))
			data := HomePageData{
				Version:       string(version),
				LatestVersion: string(latestVersion),
				Debug:         cfg.Server.Debug,
				Schedule:      cfg.Schedule,
			}
			tmpl.Execute(w, data)
		} else {
			if fileExists(path) {
				d, _ := ioutil.ReadFile(string(path))
				w.Write(d)
			} else {
				// fmt.Println(path)
				http.NotFound(w, r)
			}
		}
	})
	http.HandleFunc("/switch", switchHandler)
	http.HandleFunc("/pin", pinHandler)
	http.HandleFunc("/system", systemHandler)
	http.HandleFunc("/schedule", scheduleHandler)
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

func configureCron(schedule Schedule) {
	if cronLib != nil {
		numEntries := len(cronLib.Entries())
		for i := 0; i < numEntries; i++ {
			cronLib.Remove(cronLib.Entries()[0].ID)
		}
	} else {
		cronLib = cron.New()
	}
	cronLib.AddFunc("@every 45m", checkForUpdates)

	if schedule.OnHour != "" && schedule.OffHour != "" && schedule.OnMinutes != "" && schedule.OffMinutes != "" {
		onTime := fmt.Sprintf("%s %s * * *", schedule.OnMinutes, schedule.OnHour)
		cronLib.AddFunc(onTime, func() {
			switchLight("on")
		})
		offTime := fmt.Sprintf("%s %s * * *", schedule.OffMinutes, schedule.OffHour)
		cronLib.AddFunc(offTime, func() {
			switchLight("off")
		})
	}
	cronLib.Start()
}

func scheduleHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	op := req.FormValue("op")
	if op == "clear" {
		cfg := readConfig()
		cfg.Schedule = Schedule{}
		writeConfig(cfg)
		configureCron(cfg.Schedule)
	} else if op == "update" {
		onTime := req.FormValue("onTime")
		offTime := req.FormValue("offTime")
		if onTime == "" || offTime == "" {
			fmt.Fprintf(w, "error")
		} else {
			cfg := readConfig()
			cfg.Schedule = Schedule{
				strings.Split(onTime, ":")[0],
				strings.Split(onTime, ":")[1],
				strings.Split(offTime, ":")[0],
				strings.Split(offTime, ":")[1],
			}
			writeConfig(cfg)
			configureCron(cfg.Schedule)
		}
	} else {
		fmt.Fprintf(w, "error")
	}
}

func switchHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	op := req.FormValue("op")
	switchLight(op)
}

func switchLight(op string) {
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
		if cfg.Server.Role == "leader" && cfg.Server.FollowerIP != "" {
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
	switchLight(op)
	fmt.Fprintf(w, op)
}

func systemHandler(w http.ResponseWriter, req *http.Request) {
	op := req.FormValue("op")
	var args []string = []string{}
	var command string = ""
	if op == "shutdown" {
		command = "sudo"
		args = []string{"shutdown", "now"}
		fmt.Fprintf(w, "shutting down")
	} else if op == "reboot" {
		command = "sudo"
		args = []string{"reboot", "now"}
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
		cmd := exec.Command(command, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Start()
		if err != nil {
			fmt.Println("Failed to initiate command:", err)
			os.Exit(1)
		}
		fmt.Printf("Command output: %q\n", out.String())
	}
}

func cleanup(pin rpio.Pin, pinNumber int) {
	fmt.Println("Cleaning up pin", pinNumber)
	pin.Write(rpio.Low)
	rpio.Close()
}

func writeConfig(cfg config) {
	d, err := yaml.Marshal(&cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = ioutil.WriteFile("config.yml", d, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() config {
	viper.SetConfigName("config.yml")
	viper.AddConfigPath(currentdir())
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg := config{}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Println(err)
	}
	return cfg
}

func currentdir() (cwd string) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return cwd
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
