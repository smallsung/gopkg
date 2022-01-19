//  +build go1.13

package errors

import (
	stderrors "errors"
)

// Is 兼容标准库
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As 兼容标准库
func As(err error, target interface{}) bool { return stderrors.As(err, target) }

// Unwrap 兼容标准库
func Unwrap(err error) error {
	return unwrap(err)
}
