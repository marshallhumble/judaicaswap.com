package models

import (
	"database/sql"
	"errors"
	"net/smtp"
	"strconv"
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

func (m *ServerConfigModel) SendMail(rEmail, sEmail, item string) error {

	s, err := m.GetConfig()
	if err != nil {
		return err
	}
	server := s.mailServer + ":" + strconv.Itoa(s.mailPort)
	auth := smtp.PlainAuth("", s.mailUsername, s.mailPassword, s.mailServer)

	// Here we do it all: connect to our server, set up a message and send it

	to := []string{rEmail}

	MsgFile := []byte("To: " + rEmail + "\r\n" +
		"Subject: Email from JudaicaWebSwap about one of your items \r\n" +
		"\r\n" + sEmail + " wants to talk to you about " + item + " send an email to " + sEmail + " to " +
		"work out details!\r\n")

	if err := smtp.SendMail(server, auth, s.mailUsername, to, MsgFile); err != nil {
		return err
	}

	return nil
}
