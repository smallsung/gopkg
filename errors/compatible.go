package errors

//causer github.com/pkg/errors github.com/juju/errors
type causer interface {
	//Cause 不应该返回 nil
	Cause() error
}

func (e *errorWrap) Cause() error { return e.Unwrap() }

//underlying github.com/juju/errors
type underlying interface {
	Underlying() error
}

func (e *errorWrap) Underlying() error { return e.Unwrap() }

//Messenger github.com/juju/errors
type Messenger interface {
	Message() string
}

func (m *stringAnnotate) Messenger() string {
	return m.Annotate()
}

func unwrap(err error) error {
	switch i := err.(type) {
	case UnWrapper:
		return i.Unwrap()

	case underlying:
		// github.com/juju/errors.Wrap 不做验证
		return i.Underlying()

	case causer:
		// github.com/pkg/errors
		// github.com/juju/errors
		return i.Cause()
	default:
		return nil
	}
}

//Err github.com/juju/errors.Err
type Err = LocatorError

//NewErr github.com/juju/errors.NewErr
func NewErr(format string, args ...interface{}) Err {
	return NewLocatorError(format, args...)
}

//NewErrWithCause github.com/juju/errors.NewErrWithCause
func NewErrWithCause(other error, format string, args ...interface{}) Err {
	return NewCauserAnnotatorLocatorError(other, format, args...)
}

//ErrorStack 兼容旧代码中使用 github.com/juju/errors.ErrorStack
//
// Deprecated: 不应该被使用,应该使用 Chain
func ErrorStack(err error) string {
	return formatPlusV(err)
}

// Annotatef 兼容旧代码中使用 github.com/juju/errors.Annotatef
//
// Deprecated: 不应该被使用,应该使用 Annotate
func Annotatef(other error, format string, args ...interface{}) error {
	return annotate(other, format, args...)
}

// WithStack 兼容旧代码中使用 github.com/pkg/errors2.WithStack

// Deprecated: 不应该被使用,应该使用 Trace
func WithStack(err error) error {
	return trace(err)
}

// WithMessage 兼容旧代码中使用 github.com/pkg/errors.WithMessage
//
// Deprecated: 不应该被使用,应该使用 Annotate
func WithMessage(err error, message string, args ...interface{}) error {
	return annotate(err, message, args...)
}

// WithMessagef 兼容旧代码中使用 github.com/pkg/errors.WithMessagef
//
// Deprecated: 不应该被使用,应该使用 Annotate
func WithMessagef(err error, format string, args ...interface{}) error {
	return annotate(err, format, args...)
}

// Errorf 兼容旧代码中使用 github.com/pkg/errors.Errorf github.com/juju/errors.Errorf
//
// Deprecated: 不应该被使用,应该使用 Format
func Errorf(format string, args ...interface{}) error {
	return formatError(format, args...)
}
