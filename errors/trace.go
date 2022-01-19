package errors

import (
	"fmt"
)

//Trace 为错误添加调用栈信息
func Trace(other error) error {
	return trace(other)
}

func trace(err error) error {
	if err == nil {
		return nil
	}
	newErr := new(withStackError)
	newErr.Wrap(err)
	newErr.SetLocation(1 + 1)
	return newErr
}

type withStackError struct {
	errorWrap
	location
}

func (w *withStackError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(w))
		default:
			_, _ = fmt.Fprintf(s, "%v", w.Unwrap())
		}

	case 's':
		_, _ = fmt.Fprintf(s, "%s", w.Unwrap())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, w, fmt.Sprintf("%s", w.Unwrap().Error()))
	}
}

//Annotate 兼容"%+v"格式
func (w *withStackError) Annotate() string { return "" }
