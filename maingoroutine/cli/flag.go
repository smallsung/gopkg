package maincli

import (
	"github.com/smallsung/gopkg/errors"
	"github.com/smallsung/gopkg/maingoroutine"
	"github.com/urfave/cli/v2"
)

var (
	EnableHTTPFlag = cli.BoolFlag{
		Name:        "http.enable",
		Aliases:     nil,
		Usage:       "",
		EnvVars:     nil,
		FilePath:    "",
		Required:    false,
		Hidden:      false,
		Value:       false,
		DefaultText: "",
		Destination: nil,
		HasBeenSet:  false,
	}
	HTTPHostFlag = cli.StringFlag{
		Name:        "http.host",
		Aliases:     nil,
		Usage:       "",
		EnvVars:     nil,
		FilePath:    "",
		Required:    false,
		Hidden:      false,
		TakesFile:   false,
		Value:       "",
		DefaultText: "",
		Destination: nil,
		HasBeenSet:  false,
	}
	HTTPPortFlag = cli.UintFlag{
		Name:        "http.port",
		Aliases:     nil,
		Usage:       "",
		EnvVars:     nil,
		FilePath:    "",
		Required:    false,
		Hidden:      false,
		Value:       0,
		DefaultText: "",
		Destination: nil,
		HasBeenSet:  false,
	}
	EnableHTTPRPCFlag = cli.BoolFlag{
		Name:        "http.rpc.enable",
		Aliases:     nil,
		Usage:       "",
		EnvVars:     nil,
		FilePath:    "",
		Required:    false,
		Hidden:      false,
		Value:       false,
		DefaultText: "",
		Destination: nil,
		HasBeenSet:  false,
	}
)

type Flags struct {
	EnableHTTPFlag    cli.BoolFlag
	HTTPHostFlag      cli.StringFlag
	HTTPPortFlag      cli.UintFlag
	EnableHTTPRPCFlag cli.BoolFlag
}

func (f Flags) ToSlice() []cli.Flag {
	return []cli.Flag{
		&f.EnableHTTPFlag,
		&f.HTTPHostFlag,
		&f.HTTPPortFlag,
		&f.EnableHTTPRPCFlag,
	}
}

func Default() Flags {
	return Flags{
		EnableHTTPFlag:    EnableHTTPFlag,
		HTTPHostFlag:      HTTPHostFlag,
		HTTPPortFlag:      HTTPPortFlag,
		EnableHTTPRPCFlag: EnableHTTPRPCFlag,
	}
}

func Parse(ctx *cli.Context, config *maingoroutine.Config) error {
	if ctx.IsSet(EnableHTTPFlag.Name) {
		config.EnableHTTP = ctx.Bool(EnableHTTPFlag.Name)
	}
	if ctx.IsSet(HTTPHostFlag.Name) {
		config.HTTPHost = ctx.String(HTTPHostFlag.Name)
	}
	if ctx.IsSet(HTTPPortFlag.Name) {
		port := ctx.Uint(HTTPPortFlag.Name)
		if port > 1<<16-1 {
			return errors.Format("%s must (0, 65535]", HTTPPortFlag)
		}
		config.HTTPPort = uint16(port)
	}
	if ctx.IsSet(EnableHTTPRPCFlag.Name) {
		config.EnableHTTPRPC = ctx.Bool(EnableHTTPRPCFlag.Name)
	}
	return nil
}
