package user

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	ID       uuid.UUID
	Login    string
	password string
}

func (m *UserModel) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	m.password = string(hashedPassword)
	return nil
}

func (m *UserModel) VerifyPassword(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(m.password), []byte(password))

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
