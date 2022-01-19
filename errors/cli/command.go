package errorscli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/smallsung/gopkg/errors"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:                   "pc",
		Aliases:                nil,
		Usage:                  "报告所执行的函数的文件和行号信息",
		UsageText:              "pc 0xffffff [12345678 [....]]",
		Description:            "",
		ArgsUsage:              "",
		Category:               "",
		BashComplete:           nil,
		Before:                 nil,
		After:                  nil,
		Action:                 PC,
		OnUsageError:           nil,
		Subcommands:            nil,
		Flags:                  nil,
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "",
		CustomHelpTemplate:     "",
	}
}

func PC(ctx *cli.Context) error {
	var lines []string
	for _, s := range ctx.Args().Slice() {
		var buff []byte
		buff = append(buff, fmt.Sprintf("%s: ", s)...)
		var pc uintptr
		if u64, err := strconv.ParseUint(s, 0, 64); err == nil {
			pc = uintptr(u64)
		}
		if frame := errors.PC(pc); frame.PC != 0 && frame.File != "" {
			buff = append(buff, fmt.Sprintf("pc(%d) %s:%d: %s", frame.PC, frame.File, frame.Line, frame.Function)...)
		}
		lines = append(lines, string(buff))
	}
	fmt.Println(strings.Join(lines, "\n"))
	return nil
}
