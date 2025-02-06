package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	//Internal
	"judaicaswap.com/internal/models"
	"judaicaswap.com/ui"
)

type templateData struct {
	CurrentYear     int
	Share           models.Share
	Shares          []models.Share
	User            models.User
	Users           []models.User
	Form            any
	Flash           string
	IsAuthenticated bool
	IsUser          bool
	IsAdmin         bool
	IsGuest         bool
	CSRFToken       string
	CFSite          string
}

func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value.
	if t.IsZero() {
		return ""
	}

	// Convert the time to UTC before formatting it.
	return t.UTC().Format("02 Jan 2006 at 15:04.05")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func fileDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02012006_15_15_04_05")
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded
	// filesystem which match the pattern 'html/pages/*.gohtml'
	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we
		// want to parse.
		patterns := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		// Use ParseFS() instead of ParseFiles() to parse the template files
		// from the ui.Files embedded filesystem.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
