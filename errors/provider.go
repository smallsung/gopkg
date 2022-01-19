package errors

import (
	"fmt"
)

type LocatorError interface {
	error
	Locator

	// Formatter
	// %s	输出最内层错误消息
	// %v	输出包含注释（如果有）错误消息。 annotate: error
	// %+v	输出错误链 pc file:line: error
	fmt.Formatter

	Caller
	Callers
	Locator
	Annotator
	UnWrapper

	causer
	underlying
}

func NewLocator() Locator {
	return new(location)
}

type locatorError struct {
	error
	location
}

func (l *locatorError) Annotate() string  { return l.Error() }
func (l *locatorError) Unwrap() error     { return nil }
func (l *locatorError) Cause() error      { return nil }
func (l *locatorError) Underlying() error { return nil }
func (l *locatorError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(l))
		default:
			_, _ = fmt.Fprintf(s, "%s", l.Error())
		}
	case 's':
		_, _ = fmt.Fprintf(s, "%s", l.Error())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, l, fmt.Sprintf("%s", l.Error()))
	}
}

func NewLocatorError(format string, args ...interface{}) LocatorError {
	return &locatorError{error: fmt.Errorf(format, args...)}
}

type causerLocatorError struct {
	errorWrap
	location
}

func (c *causerLocatorError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(c))
		default:
			_, _ = fmt.Fprintf(s, "%s", c.Unwrap().Error())
		}
	case 's':
		_, _ = fmt.Fprintf(s, "%s", c.Unwrap().Error())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, c, fmt.Sprintf("%s", c.Unwrap().Error()))
	}
}

//Annotate 兼容"%+v"格式
func (c *causerLocatorError) Annotate() string { return "" }

func NewCauserLocatorError(cause error) LocatorError {
	err := new(causerLocatorError)
	err.Wrap(cause)
	return err
}

type causerAnnotatorLocatorError struct {
	errorWrap
	stringAnnotate
	location
}

func (c *causerAnnotatorLocatorError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = fmt.Fprintf(s, "%s", formatPlusV(c))
		default:
			_, _ = fmt.Fprintf(s, "%s: %v", c.Annotate(), c.Unwrap())
		}

	case 's':
		_, _ = fmt.Fprintf(s, "%s", c.Unwrap())

	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, c, fmt.Sprintf("%s", c.Unwrap().Error()))
	}
}

func NewCauserAnnotatorLocatorError(cause error, format string, args ...interface{}) LocatorError {
	err := new(causerAnnotatorLocatorError)
	err.Wrap(cause)
	err.setAnnotate(fmt.Sprintf(format, args...))
	return err
}
