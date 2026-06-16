package wallets

import "net/http"

// Error codes, mirroring the typed-error convention of sharia/errors.go.
const (
	ErrInvalidParams     = "ERR_INVALID_PARAMS"
	ErrNotFound          = "ERR_NOT_FOUND"
	ErrInsufficientFunds = "ERR_INSUFFICIENT_FUNDS"
	ErrHoldNotActive     = "ERR_HOLD_NOT_ACTIVE"
	ErrDuplicate         = "ERR_DUPLICATE"
)

type Error struct {
	Code     string `json:"error"`
	Message  string `json:"message"`
	WalletID string `json:"wallet_id,omitempty"`
	HoldID   string `json:"hold_id,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrInvalidParams:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrDuplicate:
		return http.StatusConflict
	case ErrInsufficientFunds, ErrHoldNotActive:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func newError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}
