package err

import "fmt"

type ErrKind int

const (
	Validation ErrKind = iota
	NotFound
	Fatal
)

type AppErr struct {
	msg  string
	code string
	kind ErrKind
}

func (e *AppErr) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.msg)
}

func (e *AppErr) Msg() string {
	return e.msg
}

func (e *AppErr) Code() string {
	return e.code
}

func (e *AppErr) Kind() ErrKind {
	return e.kind
}

func NewValidationErr(code, msg string) *AppErr {
	return &AppErr{code: code, msg: msg, kind: Validation}
}

func NewFatalErr(code, msg string) *AppErr {
	return &AppErr{code: code, msg: msg, kind: Fatal}
}

func NewNotFoundErr(code, msg string) *AppErr {
	return &AppErr{code: code, msg: msg, kind: NotFound}
}
