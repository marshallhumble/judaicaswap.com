package main

import (
	"errors"
	"net/http"
	"strconv"
	//Internal
	"judaicaswap.com/internal/models"
	"judaicaswap.com/internal/validator"
)

type userSignupForm struct {
	Id                  int    `form:"-"`
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	Admin               bool   `form:"admin"`
	User                bool   `form:"user"`
	Guest               bool   `form:"guest"`
	Question1           string `form:"question1"`
	Question2           string `form:"question2"`
	Question3           string `form:"question3"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userContactForm struct {
	Name    string `form:"name"`
	Email   string `form:"email"`
	Message string `form:"message"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = userSignupForm{}

	app.render(w, r, http.StatusOK, "signup.gohtml", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	if err := app.decodePostForm(r, &form); err != nil {
		data := app.newTemplateData(r)
		data.Form = form
		app.sessionManager.Put(r.Context(), "flash", "Error, please try again.")
		app.render(w, r, http.StatusUnprocessableEntity, "signup.gohtml", data)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX),
		"email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 12), "password",
		"This field must be at least 12 characters long")
	form.CheckField(validator.NotBlank(form.Question1), "question1", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Question2), "question2", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Question3), "question3", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.gohtml", data)
		return
	}

	verification := app.AlphaNumStringGen(17)

	// Try to create a new user record in the database. If the email already
	// exists then add an error message to the form and re-display it.
	//Insert(name, email, password, question1, question2, question3 string, admin, user, guest, disabled bool) error
	if err := app.users.Insert(form.Name, form.Email, form.Password, form.Question1, form.Question2, form.Question3,
		false, true, false, true, verification); err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.gohtml", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked.
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	if err := app.config.SendVerificationEmail(form.Name, form.Email, verification); err != nil {
		app.logger.Error(err.Error())
	}

	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.gohtml", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Decode the form data into the userLoginForm struct.
	var form userLoginForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)

		return
	}

	// Do some validation checks on the form. We check that both email and
	// password are provided, and also check the format of the email address as
	// a UX-nicety (in case the user makes a typo).
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.sessionManager.Put(r.Context(), "flash", "Incorrect Email or Password")
		app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		return
	}

	// Check whether the credentials are valid. If they're not, add a generic
	// non-field error message and re-display the login page.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login
	// and logout operations). -- OWASP Session Fixation Mitigation
	if err = app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	// Add the ID & Email of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "authenticatedUserEmail", form.Email)

	app.sessionManager.Put(r.Context(), "flash", "You've been logged in successfully!")

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	// Redirect the user to the application home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// func ping used for testing, just sends a reply to verify the server is up
func ping(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func (app *application) getAllUsers(w http.ResponseWriter, r *http.Request) {
	if !app.isAdmin(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	users, err := app.users.GetAllUsers()

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Users = users

	app.render(w, r, http.StatusOK, "users.gohtml", data)

}

func (app *application) editUser(w http.ResponseWriter, r *http.Request) {
	if !app.isAdmin(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	user, err := app.users.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, r, http.StatusOK, "user_edit.gohtml", data)

}

func (app *application) editUserPost(w http.ResponseWriter, r *http.Request) {
	if !app.isAdmin(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	var form userSignupForm

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	if err = app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user, err := app.users.UpdateUser(id, form.Name, form.Email, form.Password, form.Admin, form.User, form.Guest)

	data := app.newTemplateData(r)
	data.User = user

	app.sessionManager.Put(r.Context(), "flash", "Information Updated")
	app.render(w, r, http.StatusOK, "user_edit.gohtml", data)
}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !app.isAdmin(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	if err := app.users.DeleteUser(id); err != nil {
		app.serverError(w, r, err)
	}

	app.sessionManager.Put(r.Context(), "flash", "User Deleted")
	http.Redirect(w, r, "/users/", http.StatusSeeOther)
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	if !app.isAuthenticated(r) {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusOK, "home.gohtml", data)
	}

	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	user, err := app.users.Get(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	shares, err := app.Share.GetAllFromUser(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data.Shares = shares

	app.render(w, r, http.StatusOK, "user_password.gohtml", data)
}

func (app *application) updateUserPost(w http.ResponseWriter, r *http.Request) {

	var form userSignupForm

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	if err = app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user, err := app.users.UpdateUser(id, form.Name, form.Email, form.Password, form.Admin, form.User, form.Guest)

	data := app.newTemplateData(r)
	data.User = user

	app.sessionManager.Put(r.Context(), "flash", "Information Updated")
	app.render(w, r, http.StatusOK, "user_password.gohtml", data)
}

func (app *application) Contact(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userContactForm{}

	app.render(w, r, http.StatusOK, "about.gohtml", data)
}

func (app *application) ContactPost(w http.ResponseWriter, r *http.Request) {

	var form userContactForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.config.ContactFormEmail(form.Name, form.Email, form.Message); err != nil {
		app.clientError(w, http.StatusBadRequest)
		data := app.newTemplateData(r)
		data.Form = form
		app.sessionManager.Put(r.Context(), "flash", "Error try again")
		app.render(w, r, http.StatusOK, "about.gohtml", data)
	}

	data := app.newTemplateData(r)
	data.Form = form
	app.sessionManager.Put(r.Context(), "flash", "Message Sent!")
	app.render(w, r, http.StatusOK, "about.gohtml", data)
}

func (app *application) EmailVerification(w http.ResponseWriter, r *http.Request) {
	verify := r.PathValue("verify")

	passed, err := app.users.CheckVerification(verify)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if passed {
		app.sessionManager.Put(r.Context(), "flash", "Email Verified")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func (app *application) PasswordResetPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.sessionManager.Put(r.Context(), "flash", "Email needed for password reset")
		app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		return
	}

	verifyToken := app.AlphaNumStringGen(17)

	err := app.users.SetVerificationCode(form.Email, verifyToken)
	if err != nil {
		data := app.newTemplateData(r)
		data.Form = form
		app.sessionManager.Put(r.Context(), "flash", "Error sending password reset email")
		app.render(w, r, http.StatusUnprocessableEntity, "login.gohtml", data)
		app.logger.Error(err.Error())
		return
	}

	if err = app.config.SendPasswordResetEmail(form.Email, verifyToken); err != nil {
		app.logger.Error(err.Error())
	}

	app.sessionManager.Put(r.Context(), "flash", "Password reset email sent!")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

func (app *application) PasswordResetValidate(w http.ResponseWriter, r *http.Request) {

	verify := r.PathValue("verify")

	passed, err := app.users.CheckVerification(verify)

	if err != nil {
		app.logger.Error(err.Error())
		app.sessionManager.Put(r.Context(), "flash", "Linked Expired or Invalid Link")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}

	if passed {
		http.Redirect(w, r, "/user/resetPassword/"+verify, http.StatusSeeOther)
	}
}

func (app *application) PasswordResetAfterVerified(w http.ResponseWriter, r *http.Request) {
	verify := r.PathValue("verify")
	var form userLoginForm

	if err := app.decodePostForm(r, &form); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 12), "password",
		"This field must be at least 12 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = userLoginForm{}
		app.render(w, r, http.StatusOK, "password_reset.gohtml", data)
		return
	}

	if err := app.users.ResetPassword(verify, form.Password); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Password reset!")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}
