package mocks

import (
	"judaicaswap.com/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name, email, password, question1, question2, question3 string, admin, user, guest,
	disabled bool, verification string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id int) (exist bool, admin bool, user bool, guest bool, disabled bool, error error) {
	switch id {
	case 1:
		return true, false, true, false, false, nil
	default:
		return false, false, true, false, false, nil
	}
}

func (m *UserModel) GetAllUsers() ([]models.User, error) {
	return []models.User{}, nil
}
func (m *UserModel) Get(id int) (models.User, error) {
	return models.User{}, nil
}
func (m *UserModel) UpdateUser(id int, name, email, password string, admin, user, guest bool) (models.User, error) {
	return models.User{}, nil
}
func (m *UserModel) DeleteUser(id int) error {
	return nil
}
func (m *UserModel) CheckVerification(verify string) error { return nil }
