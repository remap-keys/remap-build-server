package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r)
	}
	http.HandleFunc("/", h)
	log.Printf("Remap Build Server is running.\n")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Building a QMK Firmware started.")
	cmd := exec.Command(
		"/root/.local/bin/qmk", "compile",
		"-kb", "yoichiro/lunakey_mini",
		"-km", "default")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Println("Building a QMK Firmware finished.")
	stdoutString := stdout.String()
	if err != nil {
		log.Println("Building failed.")
		stderrString := stderr.String()
		log.Printf("[ERROR] %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf("[STDOUT] %s\n", stdoutString))
		io.WriteString(w, fmt.Sprintf("[STDERR] %s\n", stderrString))
		return
	}
	log.Println("Building succeeded.")
	io.WriteString(w, stdoutString)
}
