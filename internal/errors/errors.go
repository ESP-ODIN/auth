package errors

import "errors"

var ErrorUserNotFound = errors.New("user not found")
var ErrorInvalidCredentials = errors.New("invalid credentials")
var ErrorUsernameAlreadyExists = errors.New("user with this username already exists")
var ErrorEmailAlreadyExists = errors.New("user with this email already exists")
var ErrorUserUnderage = errors.New("user must be at least 18 years old")
