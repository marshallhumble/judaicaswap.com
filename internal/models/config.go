package models

import (
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
)

type ServerConfigInterface interface {
	GetConfig() (ServerConfig, error)
	SendMail(rEmail, sEmail, fName string) error
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
