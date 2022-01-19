package devtool

import (
	"fmt"
	"runtime"

	"github.com/smallsung/gopkg/errors"
)

type Color string

const (
	Reset       Color = "\033[0m"
	Red               = "\033[31m"
	RedBold           = "\033[31;1m"
	Green             = "\033[32m"
	GreenBold         = "\033[32;1m"
	Yellow            = "\033[33m"
	Blue              = "\033[34m"
	Magenta           = "\033[35m"
	Cyan              = "\033[36m"
	White             = "\033[37m"
	BlueBold          = "\033[34;1m"
	MagentaBold       = "\033[35;1m"
	YellowBold        = "\033[33;1m"
)

func location() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}

// Print 向标注输出写入调用位置
func Print(i ...interface{}) {
	var message string
	header := fmt.Sprintf("%sLocation: %s%s\n", GreenBold, location(), Reset)
	var line string
	for i := 0; i < len(header); i++ {
		line += "-"
	}
	message = fmt.Sprintf("%s%s%s\n", GreenBold, line, Reset)
	message += fmt.Sprintf("%s%s%s\n", GreenBold, header, Reset)
	for _, ii := range i {
		message += fmt.Sprintf("%v\n", ii)
	}
	message += fmt.Sprintf("%s%s%s\n", GreenBold, line, Reset)
	fmt.Printf(message)
}

func Printf(format string, i ...interface{}) {
	var message string
	header := fmt.Sprintf("%sLocation: %s%s\n", GreenBold, location(), Reset)
	var line string
	for i := 0; i < len(header); i++ {
		line += "-"
	}
	message = fmt.Sprintf("%s%s%s\n", GreenBold, line, Reset)
	message += fmt.Sprintf("%s%s%s\n", GreenBold, header, Reset)
	message += fmt.Sprintf(format, i)
	message += fmt.Sprintf("\n%s%s%s\n", GreenBold, line, Reset)
	fmt.Printf(message)
}

func PrintError(err error) {
	var message string

	header := fmt.Sprintf("%s: Call Error Printer", location())
	var line string
	for i := 0; i < len(header); i++ {
		line += "-"
	}
	message = fmt.Sprintf("%s%s%s\n", RedBold, line, Reset)
	message += fmt.Sprintf("%s%s%s\n\n", Red, header, Reset)
	message += fmt.Sprintf("%s\n", errors.ErrorStack(err))
	message += fmt.Sprintf("%s%s%s\n", RedBold, line, Reset)
	fmt.Printf(message)
}
