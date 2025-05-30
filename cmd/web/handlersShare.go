package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	//Internal
	"judaicaswap.com/internal/models"
	"judaicaswap.com/internal/validator"
)

type shareCreateForm struct {
	ItemName            string    `form:"itemName"`
	Description         string    `form:"description"`
	Owner               int       `form:"owner"`
	Email               string    `form:"email"`
	Files               []os.File `form:"uploadFile"`
	Picture1            string    `form:"picture0"`
	Picture2            string    `form:"picture1"`
	Picture3            string    `form:"picture2"`
	Picture4            string    `form:"picture3"`
	Picture5            string    `form:"picture4"`
	PutString1          string    `form:"putString0"`
	PutString2          string    `form:"putString1"`
	PutString3          string    `form:"putString2"`
	PutString4          string    `form:"putString3"`
	PutString5          string    `form:"putString4"`
	ShipsIntl           bool      `form:"shipsIntl"`
	PayShip             bool      `form:"payShip"`
	ProdURL             string    `form:"prodUrl"`
	Avail               bool      `form:"avail"`
	Expires             int       `form:"expires"`
	validator.Validator `form:"-"`
}

// getHome Want to show all the listings but make people login to see specifics,
// so no ability to take any action here other than view
func (app *application) getHome(w http.ResponseWriter, r *http.Request) {

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

// getShareView look at a specific item, pull it from the DB, should have to authenticate
// to get access.
func (app *application) getShareView(w http.ResponseWriter, r *http.Request) {

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

// postSendMail we want to get data to send an email without exposing the email address of the owner
// gather id and sender's email from URL and session
func (app *application) postSendMail(w http.ResponseWriter, r *http.Request) {
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
	itemURL := "https://www.judaicawebswap.com/items/view/" + r.PathValue("id")

	if err := app.config.SendMail(email, sEmail, itemURL); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Email Sent!")
	http.Redirect(w, r, fmt.Sprintf("/items/view/%d", id), http.StatusSeeOther)
}

func (app *application) getShareCreate(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	data := app.newTemplateData(r)

	data.Form = &shareCreateForm{}

	app.render(w, r, http.StatusOK, "create.gohtml", data)
}

func (app *application) postShareCreate(w http.ResponseWriter, r *http.Request) {

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

	//Insert(owner int, email, title, description, produrl, picture1, picture2, picture3, picture4,
	//	picture5 string, ships, payship, avail bool, expires int) (int, error)
	id, err := app.Share.Insert(owner, email, form.ItemName, form.Description, form.ProdURL, form.Picture1, form.Picture2,
		form.Picture3, form.Picture4, form.Picture5, form.ShipsIntl, form.PayShip, true, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Item successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/items/view/%d", id), http.StatusSeeOther)
}

func (app *application) getShareEdit(w http.ResponseWriter, r *http.Request) {

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

func (app *application) postShareEdit(w http.ResponseWriter, r *http.Request) {
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

	//UpdateShare(id int, title, description, produrl, picture1, picture2, picture3, picture4,
	//	picture5 string, ships, payship, avail bool) (Share, error)
	share, err := app.Share.UpdateShare(id, form.ItemName, form.Description, form.ProdURL, form.Picture1, form.Picture2,
		form.Picture3, form.Picture4, form.Picture5, form.ShipsIntl, form.PayShip, form.Avail)

	data := app.newTemplateData(r)
	data.Share = share

	app.sessionManager.Put(r.Context(), "flash", "Item successfully updated!")
	http.Redirect(w, r, fmt.Sprintf("/items/view/%d", id), http.StatusSeeOther)
}

func (app *application) getShareDelete(w http.ResponseWriter, r *http.Request) {
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

	app.sessionManager.Put(r.Context(), "flash", "Item successfully deleted!")
	http.Redirect(w, r, fmt.Sprintf("/"), http.StatusSeeOther)
}
