package mocks

import (
	"judaicaswap.com/internal/models"
	"time"
)

var MockShares = models.Share{
	ID:          1,
	Owner:       1,
	Email:       "foo@bar.com",
	Title:       "Title Mock",
	Description: "Description for mock item",
	Picture1:    "string",
	Picture2:    "string",
	Picture3:    "string",
	Picture4:    "string",
	Picture5:    "string",
	ShipsIntl:   false,
	Available:   true,
	Created:     time.Now(),
	Expires:     time.Now().Add(time.Hour * 24 * 30),
}

type ShareModel struct{}

func (m *ShareModel) Insert(owner int, email, title, description, picture1, picture2, picture3, picture4,
	picture5 string, ships, avail bool, expires int) (int, error) {
	return 2, nil
}

func (m *ShareModel) Get(id int) (models.Share, error) {
	return models.Share{}, nil
}

func (m *ShareModel) GetEmail(id int) string {
	return "foo@bar.com"
}
func (m *ShareModel) GetAll() ([]models.Share, error) {
	return []models.Share{MockShares}, nil
}
func (m *ShareModel) Remove(id int) error {
	return nil
}
func (m *ShareModel) GetAllFromUser(id int) ([]models.Share, error) {
	return []models.Share{MockShares}, nil
}
func (m *ShareModel) UpdateShare(id int, title, description, picture1, picture2, picture3, picture4,
	picture5 string, ships, avail bool) (models.Share, error) {
	return models.Share{}, nil
}
