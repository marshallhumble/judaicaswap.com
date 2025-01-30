package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) getSignedUploadURL(w http.ResponseWriter, r *http.Request) {
	if r.Host != "localhost" {
		http.NotFound(w, r)
		return
	}

	fileString := r.PathValue("file")
	fileType := r.FormValue("ext")

	request := app.putPresignURL(fileString, fileType)
	js, err := json.Marshal(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func (app *application) hostname(w http.ResponseWriter, r *http.Request) {
	name := r.Host

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(name))

}
