package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
)

type ServerConfigInterface interface {
	GetConfig() (ServerConfig, error)
	SendMail(rEmail, sEmail, fName string) error
	ContactFormEmail(name, email, message string) error
	SendVerificationEmail(name, email, verify string) error
}

type ServerConfig struct {
	mailServer   string
	mailUsername string
	mailPassword string
	mailPort     int
	serverName   string
}

type ServerConfigModel struct {
	DB *sql.DB
}

type EmailUser struct {
	Name string
	Link string
}

func (m *ServerConfigModel) GetConfig() (ServerConfig, error) {
	stmt := `SELECT mail_server, mail_username, mail_password, mail_port, server_name FROM config`

	var c ServerConfig

	err := m.DB.QueryRow(stmt).Scan(&c.mailServer, &c.mailUsername, &c.mailPassword, &c.mailPort, &c.serverName)

	if err != nil {
		// If the query returns no rows, then row.Scan() will return a
		// sql.ErrNoRows error. We use the errors.Is() function check for that
		// error specifically, and return our own ErrNoRecord error
		// instead.
		if errors.Is(err, sql.ErrNoRows) {
			return ServerConfig{}, ErrNoRecord
		} else {
			return ServerConfig{}, err
		}
	}

	return c, nil
}

func (m *ServerConfigModel) SendMail(rEmail, sEmail, itemURL string) error {

	s, err := m.GetConfig()

	if err != nil {
		return err
	}

	subject := "Email from JudaicaWebSwap about one of your items!"
	body := fmt.Sprintf("Hi %s <br> <a href=\"mailto:%s\">%s</a> is interested in  your item on Judaica Swap! "+
		"<br> <br><a href=%s>%s</a><br><br> "+
		"Email them back at <a href=\"mailto:%s\">%s</a> <br>", rEmail, sEmail, sEmail, itemURL, itemURL, sEmail, sEmail)

	mail := gomail.NewMessage()
	mail.SetHeader("From", s.mailUsername)
	mail.SetHeader("To", rEmail)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", body)

	d := gomail.NewDialer(s.mailServer, s.mailPort, s.mailUsername, s.mailPassword)

	if err := d.DialAndSend(mail); err != nil {
		return err
	}

	return nil
}

func (m *ServerConfigModel) ContactFormEmail(name, email, message string) error {
	s, err := m.GetConfig()
	if err != nil {
		return err
	}

	message = message + "<br> email contact is: " + email + "<br>"

	mail := gomail.NewMessage()
	mail.SetHeader("From", s.mailUsername)
	mail.SetHeader("To", s.mailUsername)
	mail.SetHeader("Subject", "New contact form email from: "+name)
	mail.SetBody("text/html", message)

	d := gomail.NewDialer(s.mailServer, s.mailPort, s.mailUsername, s.mailPassword)

	if err := d.DialAndSend(mail); err != nil {
		return err
	}

	return nil
}

func (m *ServerConfigModel) SendVerificationEmail(name, email, verify string) error {
	s, err := m.GetConfig()
	if err != nil {
		return err
	}

	tmpl := `<html>
	<body>
	<h1>Hello, {{.Name}}!</h1>
	<p>Please verify your email by clicking <a href="{{.Link}}">{{.Link}}</a>.</p>
	<br>
	<p>After verification an admin will approve your account.</p>
	<p>If you did not request this, please ignore this email.</p>
	<p>Thanks,
	The Judaica Web Swap Team</p>
	</body>
	</html>`

	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		return err
	}

	link := "https://" + s.serverName + "/verify/" + verify

	user := EmailUser{Name: name, Link: link}
	var buf bytes.Buffer
	if err = t.Execute(&buf, user); err != nil {
		return err
	}

	fmt.Println(buf.String())

	mail := gomail.NewMessage()
	mail.SetHeader("From", s.mailUsername)
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", "Please verify your email address")
	mail.SetBody("text/html", buf.String())

	d := gomail.NewDialer(s.mailServer, s.mailPort, s.mailUsername, s.mailPassword)

	if err := d.DialAndSend(mail); err != nil {
		return err
	}

	return nil
}
