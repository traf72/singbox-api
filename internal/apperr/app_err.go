package apperr

type ErrKind int

const (
	Validation ErrKind = iota
	NotFound
	Conflict
	Fatal
)

type Err interface {
	error
	Msg() string
	Code() string
	Kind() ErrKind
}

type appErr struct {
	msg  string
	code string
	kind ErrKind
}

func (e *appErr) Error() string {
	return e.msg
}

func (e *appErr) Msg() string {
	return e.msg
}

func (e *appErr) Code() string {
	return e.code
}

func (e *appErr) Kind() ErrKind {
	return e.kind
}

func NewValidationErr(code, msg string) Err {
	return &appErr{code: code, msg: msg, kind: Validation}
}

func NewNotFoundErr(code, msg string) Err {
	return &appErr{code: code, msg: msg, kind: NotFound}
}

func NewConflictErr(code, msg string) Err {
	return &appErr{code: code, msg: msg, kind: Conflict}
}

func NewFatalErr(code, msg string) Err {
	return &appErr{code: code, msg: msg, kind: Fatal}
}
