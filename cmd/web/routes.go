package main

import (
	"net/http"
	"path/filepath"

	//External
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	//Create file system for static files
	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static")})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	//mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	//PING? PONG! used for testing
	mux.HandleFunc("GET /ping", ping)

	//dynamic middleware route
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	//Make Alice Login protected route
	protected := dynamic.Append(app.requireAuthentication)

	//Make Alice Admin Only route
	admin := dynamic.Append(app.requireAdmin)

	//Make an API route
	api := alice.New(app.sessionManager.LoadAndSave)

	//Default route
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.getHome))

	//Dynamic View Item (not protected)
	mux.Handle("GET /items/view/{id}", dynamic.ThenFunc(app.getShareView))

	//Dynamic User Contact Form (not protected)
	mux.Handle("GET /contact", dynamic.ThenFunc(app.getContact))
	mux.Handle("POST /contact", dynamic.ThenFunc(app.postContact))

	//Dynamic User Sign-up/Login/Logout Verify Email, Password Reset
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.getUserSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.postUserSignup))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.getUserLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.postUserLogin))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(app.postUserLogout))
	mux.Handle("GET /verify/{verify}", dynamic.ThenFunc(app.getEmailVerification))
	mux.Handle("POST /user/reset", dynamic.ThenFunc(app.postPasswordReset))
	mux.Handle("GET /user/resetPassword/{verify}", dynamic.ThenFunc(app.getPasswordResetValidate))
	mux.Handle("POST /user/resetPassword/{verify}", dynamic.ThenFunc(app.postPasswordResetAfterVerified))

	//Protected User Routes
	mux.Handle("GET /users/", admin.ThenFunc(app.getAllUsers))
	mux.Handle("GET /user/edit/{id}", protected.ThenFunc(app.getEditUser))
	mux.Handle("POST /user/edit/{id}", protected.ThenFunc(app.postEditUser))
	mux.Handle("POST /user/delete/{id}", protected.ThenFunc(app.postDeleteUser))
	mux.Handle("GET /user/update/", protected.ThenFunc(app.getUpdateUser))
	mux.Handle("POST /user/update/", protected.ThenFunc(app.postUpdateUser))

	//Protected File Create/View Routes
	mux.Handle("GET /items/create", protected.ThenFunc(app.getShareCreate))
	mux.Handle("POST /items/create", protected.ThenFunc(app.postShareCreate))
	mux.Handle("GET /items/edit/{id}", protected.ThenFunc(app.getShareEdit))
	mux.Handle("POST /items/edit/{id}", protected.ThenFunc(app.postShareEdit))
	mux.Handle("POST /items/delete/{id}", protected.ThenFunc(app.getShareDelete))
	mux.Handle("POST /items/sendEmail/{id}", protected.ThenFunc(app.postSendMail))

	//API Calls
	mux.Handle("GET /api/v1/signed/{ext}/{file}", api.ThenFunc(app.getSignedUploadURL))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
