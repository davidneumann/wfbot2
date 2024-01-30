package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/audrenbdb/goforeground"
	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/v3/process"
)

/*
	- Starts apps and return pids
	- List all running aps and pids
	- Kill app by pid
	- Get screenshot of system
	- Click at x,y position
	- Load values into clipboard
	- Send keyboard input
	- Report to master controller when it comes online
	- Have a health check endpoint master controller can use to verify it still exists
	- Issue reboot/shutdown
*/

func main() {
	r := httprouter.New()
	r.GET("/", healthCheck)
	r.GET("/apps", getApps)
	r.DELETE("/apps/:pid", killApp)
	r.POST("/apps/", startApp)
	r.POST("/apps/:pid/focus", focusApp)
	// r.GET("/screenshot", getScreenshot)
	// r.POST("/mouse", clickMouse)
	// r.POST("/clipboard", setClipboard)
	// r.POST("/keyboard", sendKeys)
	// r.POST("/shutdown", shutdown)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func healthCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "OK")
}

func getApps(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ps, err := process.Processes()
	if err != nil {
		http.Error(w, "Could not find processes", http.StatusInternalServerError)
		return
	}

	type pid struct {
		Pid  int
		Exe  string
		Name string
	}

	pids := make([]pid, len(ps))
	for _, p := range ps {
		name, err := p.Name()
		if err != nil {
			name = ""
		}
		exe, err := p.Exe()
		if err != nil {
			exe = ""
		}
		pids = append(pids, pid{Pid: int(p.Pid), Exe: exe, Name: name})
	}
	json, err := json.Marshal(pids)
	if err != nil {
		http.Error(w, "Could not find processes", http.StatusInternalServerError)
		return
	}

	fmt.Println(pids)
	fmt.Fprintf(w, string(json))
}

func killApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	targetPid, err := strconv.Atoi(params.ByName("pid"))
	if err != nil {
		http.Error(w, "Could not find processes", http.StatusInternalServerError)
		return
	}
	ps, err := process.Processes()
	if err != nil {
		http.Error(w, "Could not find processes", http.StatusInternalServerError)
		return
	}

	for _, p := range ps {
		if p.Pid == int32(targetPid) {
			p.Kill()
		}
	}
}

func startApp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)

	type Params struct {
		Path string
		Args []string
	}
	var params Params
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Println("Attempting to run", params)
	cmd := exec.Command(params.Path, params.Args...)
	err = cmd.Run()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Println("Running", cmd)
}

func focusApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	pidStr := params.ByName("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Println("Attempting to focus ", params)
	goforeground.Activate(pid)
}
