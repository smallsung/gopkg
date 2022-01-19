package maingoroutine

import (
	"context"
	"runtime"
	"strconv"

	"github.com/smallsung/gopkg/errors"
	"github.com/smallsung/gopkg/rpc"
	"github.com/smallsung/gopkg/rpc/jsonrpc"
	"go.uber.org/zap"
)

type API struct {
	Namespace string
	Service   interface{}
	Public    bool
}

func (g *Goroutine) builtinAPIs() []API {
	return []API{
		{
			Namespace: "debug",
			Service:   &debugAPI{g},
			Public:    true,
		},
	}
}

func (g *Goroutine) RegisterRPC(apis ...API) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.state != initializingState {
		return errors.New("can't register api on running/stopped main goroutine")
	}
	g.apis = append(g.apis, apis...)
	return nil
}

func (g *Goroutine) AttachRPC(ctx context.Context) *rpc.Client {
	return rpc.DialInProc(ctx, g.inProcRPCServer, jsonrpc.NewClientCodec)
}

func (g *Goroutine) configuredRPC() error {
	g.logger.Info("registering inproc rpc")
	for _, api := range g.apis {
		if err := g.inProcRPCServer.Register(api.Namespace, api.Service); err != nil {
			g.logger.Warn("register inproc rpc failure", zap.String("n", api.Namespace))
			return errors.Annotatef(err, "register inproc rpc")
		}
	}
	g.logger.Info("registered inproc rpc success")

	g.logger.Info("registering http rpc")
	if g.config.EnableHTTPRPC {
		if err := g.httpServer.enableRPC(g.apis); err != nil {
			return errors.Annotatef(err, "register http rpc")
		}
	}
	g.logger.Info("registered http rpc success")

	return nil
}

type debugAPI struct {
	g *Goroutine
}

func (api *debugAPI) Pc(hex string) (runtime.Frame, error) {
	if pc, err := strconv.ParseUint(hex, 0, 64); err != nil {
		return runtime.Frame{}, err
	} else {
		return errors.PC(uintptr(pc)), nil
	}
}
