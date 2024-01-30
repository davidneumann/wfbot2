package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/audrenbdb/goforeground"
	"github.com/go-vgo/robotgo"
	"github.com/julienschmidt/httprouter"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/vova616/screenshot"
)

func main() {
	r := httprouter.New()
	r.GET("/", healthCheck)
	r.GET("/apps", getApps)
	r.DELETE("/apps/:pid", killApp)
	r.POST("/apps/", startApp)
	r.POST("/apps/:pid/focus", focusApp)
	r.GET("/screenshot", getScreenshot)
	r.POST("/mouse/click", clickMouse)
	r.POST("/mouse/move", moveMouse)
	r.POST("/keyboard/paste", sendKeys)
	r.POST("/system/shutdown", shutdown)
	r.POST("/system/reboot", reboot)

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

func getScreenshot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	img, err := screenshot.CaptureScreen()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	w.Write(buf.Bytes())
}

func moveHandling(r *http.Request) (bool, error) {
	type Params struct {
		X int
		Y int
	}
	params := Params{
		X: -1,
		Y: -1,
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&params)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	if params.X >= 0 && params.Y >= 0 {
		fmt.Println("Moving mouse", params.X, params.Y)
		robotgo.Move(params.X, params.Y)
		time.Sleep(66 * time.Millisecond)
		return true, nil
	}

	return false, nil
}

func clickMouse(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := moveHandling(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return

	}

	fmt.Println("Clicking mouse")
	robotgo.Click("left", true)
}

func moveMouse(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	clicked, err := moveHandling(r)
	if err != nil || !clicked {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func sendKeys(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	type Params struct {
		Text *string
	}
	var params Params
	err := decoder.Decode(&params)
	if err != nil || params.Text == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fmt.Println("Pasting string", params.Text)
	robotgo.PasteStr(*params.Text)
}

func shutdown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := shutdownNow(false)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func reboot(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := shutdownNow(true)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
