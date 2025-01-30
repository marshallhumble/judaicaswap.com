package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) getSignedUploadURL(w http.ResponseWriter, r *http.Request) {
	if r.Host != "localhost" || r.Host != "judaicaswap.com" {
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
