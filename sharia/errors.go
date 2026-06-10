package sharia

import "net/http"

// Error codes (spec §10)
const (
	ErrInvalidParams     = "ERR_INVALID_PARAMS"
	ErrNotFound          = "ERR_NOT_FOUND"
	ErrInvalidTransition = "ERR_INVALID_TRANSITION"
	ErrPrecondition      = "ERR_PRECONDITION"
	ErrShariaViolation   = "ERR_SHARIA_VIOLATION"
	ErrDuplicate         = "ERR_DUPLICATE"
)

type Error struct {
	Code        string `json:"error"`
	Message     string `json:"message"`
	StandardRef string `json:"standard_ref,omitempty"`
	ContractID  string `json:"contract_id,omitempty"`
	AuditSeq    int    `json:"audit_seq,omitempty"`
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
	case ErrInvalidTransition, ErrDuplicate:
		return http.StatusConflict
	case ErrPrecondition, ErrShariaViolation:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func newError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}
