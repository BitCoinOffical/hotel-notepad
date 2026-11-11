package domain

import "errors"

var ErrNotMatch = errors.New("the data does not match")
var ErrUserConflict = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("incorrect email or password")
