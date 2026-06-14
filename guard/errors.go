package guard

import "net/http"

const (
	ErrGuardDenied = "ERR_GUARD_DENIED"
	ErrInvalidRule = "ERR_INVALID_RULE"
	ErrNotFound    = "ERR_NOT_FOUND"
)

type Error struct {
	Code        string `json:"error"`
	Message     string `json:"message"`
	StandardRef string `json:"standard_ref,omitempty"`
	RuleID      string `json:"rule_id,omitempty"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

func (e *Error) HTTPStatus() int {
	switch e.Code {
	case ErrInvalidRule:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrGuardDenied:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func newError(code, message string) *Error { return &Error{Code: code, Message: message} }
