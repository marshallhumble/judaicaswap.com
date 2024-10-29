package main

import (
	"errors"
	"fmt"
	"github.com/google/safeopen"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	//Internal
	"judaicaswap.com/internal/models"
	"judaicaswap.com/internal/validator"
)

type shareCreateForm struct {
	ItemName            string    `form:"itemName"`
	Description         string    `form:"description"`
	Owner               int       `form:"-"`
	Email               string    `form:"email"`
	Files               []os.File `form:"uploadFile"`
	Picture1            string    `form:"-"`
	Picture2            string    `form:"-"`
	Picture3            string    `form:"-"`
	Picture4            string    `form:"-"`
	Picture5            string    `form:"-"`
	ShipsIntl           bool      `form:"-"`
	Avail               bool      `form:"avail"`
	Expires             int       `form:"expires"`
	validator.Validator `form:"-"`
}

// home Want to show all the listings but make people login to see specifics,
// so no ability to take any action here other than view
func (app *application) home(w http.ResponseWriter, r *http.Request) {

	shares, err := app.Share.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Shares = shares

	app.render(w, r, http.StatusOK, "home.gohtml", data)
	return

}

// shareView look at a specific item, pull it from the DB, should have to authenticate
// to get access.
func (app *application) shareView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	sharedItem, err := app.Share.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Share = sharedItem

	app.render(w, r, http.StatusOK, "view.gohtml", data)
}

// sendMail we want to get data to send an email without exposing the email address of the owner
// gather id and sender's email from URL and session
func (app *application) sendMail(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	email := app.Share.GetEmail(id)
	sEmail := app.sessionManager.GetString(r.Context(), "authenticatedUserEmail")
	itemURL := "http://localhost:4000/items/view/" + r.PathValue("id")

	if err := app.config.SendMail(email, sEmail, itemURL); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Email Sent!")
	http.Redirect(w, r, fmt.Sprintf("/items/view/%d", id), http.StatusSeeOther)
}

func (app *application) shareCreate(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	data := app.newTemplateData(r)

	data.Form = &shareCreateForm{}

	app.render(w, r, http.StatusOK, "create.gohtml", data)
}

func (app *application) shareCreatePost(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	var form shareCreateForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	multipartFormData := r.MultipartForm

	for key, file := range multipartFormData.File["uploadFile"] {

		if key < 5 {

			ext := filepath.Ext(file.Filename)
			nTime := fileDate(time.Now())
			id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

			file.Filename = strings.ReplaceAll(strconv.Itoa(id)+strings.ToLower(strings.TrimSuffix(file.Filename,
				filepath.Ext(file.Filename))), " ", "-") + "-" + fmt.Sprintf("%v", nTime) + ext

			dst, err := safeopen.CreateAt("./ui/static/SharePics", file.Filename)
			if err != nil {
				app.clientError(w, http.StatusBadRequest)
				return
			}

			f, _ := file.Open()
			io.Copy(dst, f)

			switch key {
			case 0:
				form.Picture1 = file.Filename
			case 1:
				form.Picture2 = file.Filename
			case 2:
				form.Picture3 = file.Filename
			case 3:
				form.Picture4 = file.Filename
			case 4:
				form.Picture5 = file.Filename
			default:
				form.Picture1 = file.Filename
			}
		}
	}

	form.CheckField(validator.NotBlank(form.Description),
		"description", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.ItemName),
		"itemName", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "create.gohtml", data)
		return
	}

	owner := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	email := app.sessionManager.GetString(r.Context(), "authenticatedUserEmail")

	//Insert(owner int, email, title, description, picture1, picture2, picture3, picture4,
	//		picture5 string, ships, avail bool, expires int16) (int, error)
	id, err := app.Share.Insert(owner, email, form.ItemName, form.Description, form.Picture1, form.Picture2,
		form.Picture3, form.Picture4, form.Picture5, form.ShipsIntl, true, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Item successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/items/view/%d", id), http.StatusSeeOther)
}

func (app *application) shareEdit(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	data := app.newTemplateData(r)

	data.Form = &shareCreateForm{}

	share, err := app.Share.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Share = share
	app.render(w, r, http.StatusUnprocessableEntity, "share_edit.gohtml", data)
	return
}

func (app *application) shareEditPost(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var form shareCreateForm
	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	multipartFormData := r.MultipartForm

	for key, file := range multipartFormData.File["uploadFile"] {

		if key < 5 {
			ext := filepath.Ext(file.Filename)
			nTime := fileDate(time.Now())
			id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

			file.Filename = strings.ReplaceAll(strconv.Itoa(id)+strings.ToLower(strings.TrimSuffix(file.Filename,
				filepath.Ext(file.Filename))), " ", "-") + "-" + fmt.Sprintf("%v", nTime) + ext

			dst, err := safeopen.CreateAt("./ui/static/SharePics", file.Filename)
			if err != nil {
				app.clientError(w, http.StatusBadRequest)
				return
			}

			f, _ := file.Open()
			io.Copy(dst, f)

			switch key {
			case 0:
				form.Picture1 = file.Filename
			case 1:
				form.Picture2 = file.Filename
			case 2:
				form.Picture3 = file.Filename
			case 3:
				form.Picture4 = file.Filename
			case 4:
				form.Picture5 = file.Filename
			default:
				form.Picture1 = file.Filename
			}
		}
	}

	//UpdateShare(id, title, description, picture1, picture2, picture3, picture4,
	//	picture5 string, ships, avail bool) error
	if err := app.Share.UpdateShare(id, form.ItemName, form.Description, form.Picture1, form.Picture2,
		form.Picture3, form.Picture4, form.Picture5, form.ShipsIntl, form.Avail); err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) shareDelete(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	if err := app.Share.Remove(id); err != nil {
		app.serverError(w, r, err)
	}
}
