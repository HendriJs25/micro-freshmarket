package apperror

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("resource not found")
	ErrAlreadyExists   = errors.New("resource already exists")
	ErrMisconfigured   = errors.New("application misconfigured")
	ErrConflict        = errors.New("resource conflict")

	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotVerified = errors.New("account not verified")
)
