package errors

import "errors"

var ErrorUserNotFound = errors.New("user not found")
var ErrorInvalidCredentials = errors.New("invalid credentials")
var ErrorUserAlreadyExists = errors.New("user with this email already exists")
