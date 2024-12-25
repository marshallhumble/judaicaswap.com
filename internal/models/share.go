package models

import (
	"database/sql"
	"errors"
	"time"
)

type ShareInterface interface {
	Insert(owner int, email, title, description, produrl, picture1, picture2, picture3, picture4,
		picture5 string, ships, payship, avail bool, expires int) (int, error)
	Get(id int) (Share, error)
	GetEmail(id int) string
	GetAll() ([]Share, error)
	Remove(id int) error
	GetAllFromUser(id int) ([]Share, error)
	UpdateShare(id int, title, description, produrl, picture1, picture2, picture3, picture4,
		picture5 string, ships, payship, avail bool) (Share, error)
}

type Share struct {
	ID          int
	Owner       int
	Email       string
	Title       string
	Description string
	ProdURL     string
	Picture1    string
	Picture2    string
	Picture3    string
	Picture4    string
	Picture5    string
	ShipsIntl   bool
	PayShip     bool
	Available   bool
	Created     time.Time
	Expires     time.Time
}

type ShareModel struct {
	DB *sql.DB
}

func (m *ShareModel) Insert(owner int, email, title, description, produrl, picture1, picture2, picture3, picture4,
	picture5 string, ships, payship, avail bool, expires int) (int, error) {
	stmt := `INSERT INTO shares (owner, email, title, description, produrl, picture1, picture2, picture3, picture4,
                    picture5, shipsIntl, payship, available, createdAt, expires) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, owner, email, title, description, produrl, picture1, picture2, picture3, picture4,
		picture5, ships, payship, avail, expires)
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

	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description, &s.ProdURL,
		&s.Picture1, &s.Picture2, &s.Picture3, &s.Picture4, &s.Picture5, &s.ShipsIntl, &s.PayShip, &s.Available,
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
		err := rows.Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description, &s.ProdURL, &s.Picture1, &s.Picture2,
			&s.Picture3, &s.Picture4, &s.Picture5, &s.Available, &s.ShipsIntl, &s.PayShip, &s.Created, &s.Expires)
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
		if err := rows.Scan(&s.ID, &s.Owner, &s.Email, &s.Title, &s.Description, &s.ProdURL, &s.Picture1, &s.Picture2,
			&s.Picture3, &s.Picture4, &s.Picture5, &s.ShipsIntl, &s.PayShip, &s.Available, &s.Created, &s.Expires); err != nil {
			return nil, err
		}

		shares = append(shares, s)
	}

	return shares, nil
}

func (m *ShareModel) UpdateShare(id int, title, description, produrl, picture1, picture2, picture3, picture4,
	picture5 string, ships, payship, avail bool) (Share, error) {

	var s Share

	if picture1 == "" {
		stmt := `UPDATE shares
				SET 
					title = COALESCE(?, title),
					description = COALESCE(?, description),
					produrl = COALESCE(?, produrl),
					shipsintl = COALESCE(?, shipsintl),
					shipsintl = COALESCE(?, shipsintl),
					available = COALESCE(?, available)
				WHERE id = ?;`

		_, err := m.DB.Exec(stmt, title, description, produrl, ships, payship, avail, id)

		if err != nil {
			return s, err
		}

		s.ID = id
		s.Title = title
		s.Description = description
		s.ShipsIntl = ships
		s.Available = avail

		return s, nil
	}

	stmt := `UPDATE shares
				SET 
					title = COALESCE(?, title),
					description = COALESCE(?, description),
					produrl = COALESCE(?, produrl),
					picture1 = COALESCE(?, picture1),
					picture2 = COALESCE(?, picture2),
					picture3 = COALESCE(?, picture3),
					picture4 = COALESCE(?, picture4),
					picture5 = COALESCE(?, picture5),
					shipsintl = COALESCE(?, shipsintl),
					payship = COALESCE(?, payship),
					available = COALESCE(?, available)
				WHERE id = ?;`

	_, err := m.DB.Exec(stmt, title, description, produrl, picture1, picture2, picture3, picture4, picture5, ships,
		payship, avail, id)

	if err != nil {
		return s, err
	}

	s.ID = id
	s.Title = title
	s.Description = description
	s.ProdURL = produrl
	s.Picture1 = picture1
	s.Picture2 = picture2
	s.Picture3 = picture3
	s.Picture4 = picture4
	s.Picture5 = picture5
	s.ShipsIntl = ships
	s.PayShip = payship
	s.Available = avail

	return s, nil
}
