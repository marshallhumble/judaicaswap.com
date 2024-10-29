package models

import (
	"database/sql"
	"errors"
	"time"
)

type ShareInterface interface {
	Insert(owner int, email, title, description, picture1, picture2, picture3, picture4,
		picture5 string, ships, avail bool, expires int) (int, error)
	Get(id int) (Share, error)
	GetEmail(id int) string
	GetAll() ([]Share, error)
	Remove(id int) error
	GetAllFromUser(id int) ([]Share, error)
	UpdateShare(id int, title, description, picture1, picture2, picture3, picture4,
		picture5 string, ships, avail bool) error
}

type Share struct {
	ID          int
	Owner       int
	Email       string
	Title       string
	Description string
	Picture1    string
	Picture2    string
	Picture3    string
	Picture4    string
	Picture5    string
	ShipsIntl   bool
	Available   bool
	Created     time.Time
	Expires     time.Time
}

type ShareModel struct {
	DB *sql.DB
}

func (m *ShareModel) Insert(owner int, email, title, description, picture1, picture2, picture3, picture4,
	picture5 string, ships, avail bool, expires int) (int, error) {
	stmt := `INSERT INTO shares (owner, email, title, description, picture1, picture2, picture3, picture4, picture5, 
                    shipsIntl, available, createdAt, expires) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, owner, email, title, description, picture1, picture2, picture3, picture4,
		picture5, ships, avail, expires)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {

		return 0, err
	}

	return int(id), nil

}

func (m *ShareModel) Get(id int) (Share, error) {
	stmt := `SELECT * FROM shares WHERE Expires > UTC_TIMESTAMP() AND id = ?`

	var s Share

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description,
		&s.Picture1, &s.Picture2, &s.Picture3, &s.Picture4, &s.Picture5, &s.ShipsIntl, &s.Available,
		&s.Created, &s.Expires)

	if errors.Is(err, sql.ErrNoRows) {
		return Share{}, ErrNoRecord
	} else if err != nil {
		return Share{}, err
	}

	return s, nil
}

func (m *ShareModel) GetAll() ([]Share, error) {
	stmt := `SELECT * FROM shares WHERE Expires > UTC_TIMESTAMP()`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var shares []Share

	for rows.Next() {
		var s Share
		err := rows.Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description, &s.Picture1, &s.Picture2, &s.Picture3,
			&s.Picture4, &s.Picture5, &s.Available, &s.ShipsIntl, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		shares = append(shares, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shares, nil
}

func (m *ShareModel) GetEmail(id int) string {
	stmt := `SELECT email FROM shares WHERE id = ?`
	var email string
	err := m.DB.QueryRow(stmt, id).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return ""
	}
	return email
}

func (m *ShareModel) Remove(id int) error {
	stmt := `DELETE FROM shares WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *ShareModel) GetAllFromUser(id int) ([]Share, error) {

	var shares []Share

	stmt := `SELECT * FROM shares WHERE Expires > UTC_TIMESTAMP() AND owner = ?`
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var s Share
		if err := rows.Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description, &s.Picture1, &s.Picture2, &s.Picture3,
			&s.Picture4, &s.Picture5, &s.ShipsIntl, &s.Available, &s.Created, &s.Expires); err != nil {
			return nil, err
		}

		shares = append(shares, s)
	}

	return shares, nil
}

func (m *ShareModel) UpdateShare(id int, title, description, picture1, picture2, picture3, picture4,
	picture5 string, ships, avail bool) error {
	stmt := `UPDATE shares SET title = ?, description = ?, picture1 =?, picture2 = ?, picture3 = ?, picture4 =?, 
                  picture5 = ?, shipsintl = ?, available = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, title, description, picture1, picture2, picture3, picture4, picture5, ships, avail, id)

	if err != nil {
		return err
	}

	return nil
}
