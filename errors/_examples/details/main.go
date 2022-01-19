package main

import (
	"fmt"
	"io"

	"github.com/smallsung/gopkg/errors"
)

func main() {
	err := io.EOF
	err = errors.Trace(err)
	err = errors.Annotate(err, "注释信息1")
	err = errors.Trace(err)
	err = errors.Annotate(err, "注释信息2")
	fmt.Println(errors.Details(err))
}
