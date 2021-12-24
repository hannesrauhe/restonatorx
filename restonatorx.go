package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/hannesrauhe/freeps"
)

func ExecHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cmd := exec.Command("./scripts/"+vars["script"], vars["arg"])
	stdout, err := cmd.Output()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Executed: %v\nParameters: %v\nError: %v", vars["script"], vars["arg"], string(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Executed: %v\nParameters: %v\nOutput: %v", vars["script"], vars["arg"], string(stdout))
	}
}

func FritzHandler(w http.ResponseWriter, r *http.Request) {
	f, _ := freeps.NewFreeps("/home/pi/.fritzflux/config.json")

	vars := mux.Vars(r)
	fn := vars["function"]
	if fn == "getdevicelistinfos" {
		devl, err := f.GetDeviceList()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonbytes, err := json.MarshalIndent(devl, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(jsonbytes)
		return
	}

	dev := vars["device"]
	arg := make(map[string]string)
	for key, value := range r.URL.Query() {
		arg[key] = value[0]
	}
	var err error
	if fn == "wakeup" {
		log.Printf("Waking Up %v", dev)
		err = f.WakeUpDevice(dev)
	} else {
		err = f.HomeAutoSwitch(fn, dev, arg)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "FritzHandler\nParameters: %v\nError: %v", vars, string(err.Error()))
		return
	}
	fmt.Fprintf(w, "Fritz: %v, %v, %v", fn, dev, arg)
}

func DenonHandler(w http.ResponseWriter, r *http.Request) {
	denon_address := "192.168.170.26"
	c := http.Client{}
	vars := mux.Vars(r)
	var cmd string

	switch vars["function"] {
	case "on":
		cmd = "PutSystem_OnStandby/ON"
	case "off":
		cmd = "PutSystem_OnStandby/STANDBY"
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data_url := "http://" + denon_address + "/MainZone/index.put.asp"
	data := url.Values{}
	data.Set("cmd0", cmd)

	data_resp, err := c.PostForm(data_url, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "DenonHandler\nParameters: %v\nError: %v", vars, string(err.Error()))
		return
	}
	fmt.Fprintf(w, "Denon: %v, %v", vars, data_resp)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/exec/{script:[a-z0-9_]+}/{arg:[a-z0-9_]+}", ExecHandler)
	r.HandleFunc("/script/{script:[a-z0-9_]+}/{arg:[a-z0-9_]+}", ExecHandler)
	r.HandleFunc("/fritz/{function}", FritzHandler)
	r.HandleFunc("/fritz/{function}/{device}", FritzHandler)
	r.HandleFunc("/denon/{function}", DenonHandler)

	log.Println("Starting Server")
	log.Fatal(http.ListenAndServe(":8000", r))
}
