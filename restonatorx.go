package main

import (
	"fmt"
	"log"
	"net/http"
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
	fmt.Fprintf(w, "Fritz: %v, %v", vars["function"], vars["device"])
	if vars["function"] == "wakeup" {
		f.WakeUpDevice(vars["device"])
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/exec/{script:[a-z0-9_]+}/{arg:[a-z0-9_]+}", ExecHandler)
	r.HandleFunc("/script/{script:[a-z0-9_]+}/{arg:[a-z0-9_]+}", ExecHandler)
	r.HandleFunc("/fritz/{function}/{device:[a-z0-9]+}", FritzHandler)

	log.Println("Startinng Server")
	log.Fatal(http.ListenAndServe(":8000", r))
}
