package errors

import (
	"fmt"
)

type errorString struct {
	location
	s string
}

func (e *errorString) Error() string { return e.s }

func (e *errorString) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(e))
		default:
			_, _ = fmt.Fprintf(s, "%s", e.s)
		}

	case 's':
		_, _ = fmt.Fprintf(s, "%s", e.s)

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, e, fmt.Sprintf("%s", e.Error()))
	}
}

type errorWrap struct {
	error
}

func (e *errorWrap) Unwrap() error { return e.error }

func (e *errorWrap) Wrap(err error) { e.error = err }

// New 是标准库的替代品.它返回一个包含调用栈信息的错误.
// 	New()
// 	New(message string)
// 	New(format string, args ...interface{})
func New(args ...interface{}) error {
	switch len(args) {
	case 0:
		err := new(positionError)
		err.setLocation(1)
		return err
	case 1:
		err := new(errorString)
		err.s = args[0].(string)
		err.SetLocation(1)
		return err
	default:
		err := new(errorString)
		err.s = fmt.Sprintf(args[0].(string), args[1:]...)
		err.SetLocation(1)
		return err
	}
}

// Format 如 fmt.Errorf 一样.但是返回的错误包含调用栈信息.
func Format(format string, args ...interface{}) error {
	return formatError(format, args...)
}

func formatError(format string, args ...interface{}) error {
	err := new(errorString)
	err.s = fmt.Sprintf(format, args...)
	err.SetLocation(1 + 1)
	return err
}
