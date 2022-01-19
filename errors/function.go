package errors

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Cause 通过循环 Unwrap ,它将找到最初的错误.
// 除非提供的 err 是 nil, 否则 Unwrap 不应该返回 nil
func Cause(err error) error {
	for err != nil {
		switch i := err.(type) {
		case UnWrapper:
			err = i.Unwrap()
		case causer:
			err = i.Cause()
		default:
			return err
		}
	}
	return err
}

// Chain 会循环 Unwrap ,找到 error 的每层包装, 并格式化(包含调用栈信息和注释信息)到单行
//  go run ./errors/_examples/chain/main.go
//                                                               : EOF
//  4980726 .../smallsung/gopkg/errors/_examples/chain/main.go:12:
//  4980741 .../smallsung/gopkg/errors/_examples/chain/main.go:13: 注释信息1
//  4980805 .../smallsung/gopkg/errors/_examples/chain/main.go:14:
//  4980830 .../smallsung/gopkg/errors/_examples/chain/main.go:15: 注释信息2
func Chain(err error) string {
	return formatPlusV(err)
}

func formatPlusV(err error) string {
	type stackLine struct {
		caller   string
		location string
		message  string
	}

	var lines []string

	var stackLines []*stackLine

	var callerWidth, locationWidth int

	for err != nil {
		sl := new(stackLine)
		if i, ok := err.(Caller); ok {
			sl.caller = strconv.FormatUint(uint64(i.Caller()), 10)
			if width := utf8.RuneCountInString(sl.caller); width > callerWidth {
				callerWidth = width
			}
		}

		if i, ok := err.(Location); ok {
			if file, line := i.Location(); file != "" && line != 0 {
				sl.location = fmt.Sprintf("%s:%d", file, line)
				if width := utf8.RuneCountInString(sl.location); width > locationWidth {
					locationWidth = width
				}
			}
		}

		switch i := err.(type) {
		case Annotator:
			sl.message = i.Annotate()

		case Messenger:
			sl.message = i.Message()

			//hook github.com/juju/errors
			if i, ok := err.(underlying); ok {
				underlying := i.Underlying()
				cause := Cause(err)
				if cause != nil && !reflect.DeepEqual(Cause(underlying), cause) {
					if sl.message != "" {
						sl.message += ": "
					}
					sl.message += cause.Error()
				}
			}

		default:
			sl.message = i.Error()
		}

		err = unwrap(err)
		stackLines = append(stackLines, sl)
	}

	format := fmt.Sprintf("%%%ds %%%ds: %%s", callerWidth, locationWidth)

	for i := len(stackLines) - 1; i >= 0; i-- {
		lines = append(lines, fmt.Sprintf(format, stackLines[i].caller, stackLines[i].location, stackLines[i].message))
	}

	return strings.Join(lines, "\n")
}

//Stacks 返回最内部实现 Callers 的 error 的堆栈
func Stacks(err error) string {
	var c Callers
	for err != nil {
		if i, ok := err.(Callers); ok {
			c = i
		}
		err = unwrap(err)
	}
	var lines []string
	if c != nil {
		frames := runtime.CallersFrames(c.Callers())
		for {
			f, more := frames.Next()
			lines = append(lines, fmt.Sprintf("%d %s:%d: %s", f.PC, f.File, f.Line, f.Function))
			if !more {
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

// Details 展示完整的错误信息. 它包含 最初错误信息,错误包装信息,最内部实现 Callers 的 error 的堆栈
//  go run ./errors/_examples/details/main.go
//  error:
//  EOF
//
//  errors:
//                                                                              : EOF
//  7867990 .../smallsung/gopkg/errors/_examples/details/main.go:12:
//  7868005 .../smallsung/gopkg/errors/_examples/details/main.go:13: 注释信息1
//  7868069 .../smallsung/gopkg/errors/_examples/details/main.go:14:
//  7868094 .../smallsung/gopkg/errors/_examples/details/main.go:15: 注释信息2
//
//  stacks:
//  7867990 .../smallsung/gopkg/errors/_examples/details/main.go:12: main.main
//  7377781 .../src/runtime/proc.go:225: runtime.main
//  7563840 .../src/runtime/asm_amd64.s:1371: runtime.goexit
func Details(err error) string {
	if err == nil {
		return "nil err"
	}
	var lines []string

	cause := Cause(err)
	lines = append(lines, "error:")
	lines = append(lines, cause.Error())
	lines = append(lines, "")

	lines = append(lines, "errors:")
	lines = append(lines, fmt.Sprintf("%+v", err))
	lines = append(lines, "")

	lines = append(lines, "stacks:")
	lines = append(lines, Stacks(err))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}
