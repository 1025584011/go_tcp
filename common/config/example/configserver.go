package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("/tmp/")))
	http.HandleFunc("/config", getfunc)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func getfunc(w http.ResponseWriter, r *http.Request) {
	fname := r.FormValue("name")
	file, err := os.Open("/tmp/" + fname)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	text, _ := ioutil.ReadAll(file)
	w.Write([]byte(text))
}
