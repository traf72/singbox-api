package apperr

type ErrKind int

const (
	Validation ErrKind = iota
	NotFound
	Fatal
)

type Err struct {
	msg  string
	code string
	kind ErrKind
}

func (e *Err) Error() string {
	return e.msg
}

func (e *Err) Msg() string {
	return e.msg
}

func (e *Err) Code() string {
	return e.code
}

func (e *Err) Kind() ErrKind {
	return e.kind
}

func NewValidationErr(code, msg string) *Err {
	return &Err{code: code, msg: msg, kind: Validation}
}

func NewFatalErr(code, msg string) *Err {
	return &Err{code: code, msg: msg, kind: Fatal}
}

func NewNotFoundErr(code, msg string) *Err {
	return &Err{code: code, msg: msg, kind: NotFound}
}
