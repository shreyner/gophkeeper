package user

import "errors"

var ErrLoginAlreadyExist = errors.New("login already exist")

var ErrUserNotFound = errors.New("not found")
