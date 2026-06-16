package payments

import "net/http"

// Error codes, following the typed-error convention of sharia/errors.go.
const (
	ErrUnknownPSP        = "ERR_UNKNOWN_PSP"
	ErrInvalidSignature  = "ERR_INVALID_SIGNATURE"
	ErrInvalidPayload    = "ERR_INVALID_PAYLOAD"
	ErrInvalidParams     = "ERR_INVALID_PARAMS"
	ErrInsufficientFunds = "ERR_INSUFFICIENT_FUNDS"
	ErrPSPUnavailable    = "ERR_PSP_UNAVAILABLE"
	ErrNotFound          = "ERR_NOT_FOUND"
)

type Error struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	PSP     string `json:"psp,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrInvalidParams, ErrInvalidPayload:
		return http.StatusBadRequest
	case ErrInvalidSignature:
		return http.StatusUnauthorized
	case ErrUnknownPSP, ErrNotFound:
		return http.StatusNotFound
	case ErrInsufficientFunds:
		return http.StatusUnprocessableEntity
	case ErrPSPUnavailable:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func newError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}
