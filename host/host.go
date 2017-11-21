package main

import (
	"encoding/json"
	"github.com/clear-code/mcd-go"
	"github.com/lhside/chrome-go"
	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type RequestParams struct {
	// launch
	Path string   `json:path`
	Args []string `json:args`
	Url  string   `json:url`
}
type Request struct {
	Command string        `json:"command"`
	Params  RequestParams `json:"params"`
	Logging bool          `json:"logging"`
}

var DebugLogs []string

func main() {
	log.SetOutput(ioutil.Discard)

	rawRequest, err := chrome.Receive(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	request := &Request{}
	if err := json.Unmarshal(rawRequest, request); err != nil {
		log.Fatal(err)
	}

	if request.Logging {
		logfile, err := os.OpenFile("./log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			logfilePath := os.ExpandEnv(`${temp}\com.clear_code.launchacrotray_we.log.txt`)
			logfile, err = os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
		defer logfile.Close()
		log.SetOutput(logfile)
		log.SetFlags(log.Ldate | log.Ltime)
	}

	LogForDebug("Command is " + request.Command)
	switch command := request.Command; command {
	case "launch":
		if FindAcrotrayProcess() {
			LogForDebug("acrotray.exe is already running")
		} else {
			Launch(request.Params.Path, request.Params.Args, request.Params.Url)
		}
	case "get-acrotray-path":
		SendAcrotrayPath()
	case "read-mcd-configs":
		SendMCDConfigs()
	default: // just echo
		err = chrome.Post(rawRequest, os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func LogForDebug(message string) {
	DebugLogs = append(DebugLogs, message)
	log.Print(message + "\r\n")
}

type LaunchResponse struct {
	Success bool     `json:"success"`
	Path    string   `json:"path"`
	Args    []string `json:"args"`
	Logs    []string `json:"logs"`
}

func FindAcrotrayProcess() bool {
	found := false
	processes, err := ps.Processes()
	if err != nil {
		log.Fatal(err)
	}
	for _, process := range processes {
		if process.Executable() == "acrotray.exe" {
			found = true
			break
		}
	}
	return found
}

func Launch(path string, defaultArgs []string, url string) {
	args := append(defaultArgs, url)
	command := exec.Command(path, args...)
	// "0x01000000" is the raw version of "CREATE_BREAKAWAY_FROM_JOB".
	// See also:
	//   https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Native_messaging#Closing_the_native_app
	//   https://msdn.microsoft.com/en-us/library/windows/desktop/ms684863(v=vs.85).aspx
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x01000000}
	response := &LaunchResponse{true, path, args, DebugLogs}

	err := command.Start()
	if err != nil {
		LogForDebug("Failed to launch " + path)
		log.Fatal(err)
		response.Success = false
	}
	// Wait until the launcher completely finishes.
	time.Sleep(3 * time.Second)

	response.Logs = DebugLogs
	body, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	err = chrome.Post(body, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

type SendAcrotrayPathResponse struct {
	Path string   `json:"path"`
	Logs []string `json:"logs"`
}

func SendAcrotrayPath() {
	path := GetAcrotrayPath()
	response := &SendAcrotrayPathResponse{path, DebugLogs}
	body, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	err = chrome.Post(body, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAcrotrayPath() (path string) {
	keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\Acrobat.exe`
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		keyPath,
		registry.QUERY_VALUE)
	if err != nil {
		LogForDebug("Failed to open key " + keyPath)
		log.Fatal(err)
	}
	defer key.Close()

	acrobatPath, _, err := key.GetStringValue("Path")
	if err != nil {
		LogForDebug("Failed to get value from <Path> of key " + keyPath)
		log.Fatal(err)
	}
	path = filepath.Join(acrobatPath, "acrotray.exe")
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		LogForDebug("Failed to stat acrotray.exe at " + path)
		log.Fatal(err)
	}
	return
}

type SendMCDConfigsResponse struct {
	AcrotrayApp   string   `json:"acrotrayapp,omitempty"`
	AcrotrayArgs  string   `json:"acrotrayargs,omitempty"`
	WatchInterval uint     `json:"watchinterval,omitempty"`
	Debug         bool     `json:"debug,omitempty"`
	Logs          []string `json:"logs"`
}

func SendMCDConfigs() {
	configs, err := mcd.New()
	if len(mcd.DebugLogs) > 0 {
		LogForDebug("Logs from mcd:\n  " + strings.Join(mcd.DebugLogs, "\n  "))
	}
	if err != nil {
		LogForDebug("Failed to read MCD configs.\n" + err.Error())
		//log.Fatal(err)
	}

	response := &SendMCDConfigsResponse{}

	acrotrayApp, err := configs.GetStringValue("extensions.launchacrotray.acrotrayapp")
	if err == nil {
		response.AcrotrayApp = acrotrayApp
	}
	acrotrayArgs, err := configs.GetStringValue("extensions.launchacrotray.acrotrayargs")
	if err == nil {
		response.AcrotrayArgs = acrotrayArgs
	}
	watchInterval, err := configs.GetIntegerValue("extensions.launchacrotray.watchinterval")
	if err == nil {
		if watchInterval < 0 {
			response.WatchInterval = 10
		} else {
			response.WatchInterval = uint(watchInterval)
		}
	}
	debug, err := configs.GetBooleanValue("extensions.launchacrotray.debug")
	if err == nil {
		response.Debug = debug
	}

	if len(configs.DebugLogs) > 0 {
		LogForDebug("Logs from mcd configs:\n  " + strings.Join(configs.DebugLogs, "\n  "))
	}
	response.Logs = DebugLogs
	body, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}
	err = chrome.Post(body, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
