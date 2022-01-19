package errors

import (
	"fmt"
)

// Annotate 为错误添加注释信息(包含调用栈信息)
func Annotate(other error, message string, args ...interface{}) error {
	return annotate(other, message, args...)
}

func annotate(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	newErr := new(withAnnotateError)
	newErr.Wrap(err)
	newErr.setAnnotate(fmt.Sprintf(format, args...))
	newErr.SetLocation(1 + 1)
	return newErr
}

type stringAnnotate string

func (m *stringAnnotate) setAnnotate(message string) {
	*m = stringAnnotate(message)
}

func (m *stringAnnotate) Annotate() string {
	return string(*m)
}

type withAnnotateError struct {
	errorWrap
	location
	stringAnnotate
}

func (e *withAnnotateError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(e))
		default:
			_, _ = fmt.Fprintf(s, "%s: %v", e.Annotate(), e.Unwrap())
		}

	case 's':
		_, _ = fmt.Fprintf(s, "%s", e.Unwrap())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, e, fmt.Sprintf("%s", e.Unwrap().Error()))
	}
}
