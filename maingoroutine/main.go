package maingoroutine

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/smallsung/gopkg/errors"
	"github.com/smallsung/gopkg/rpc"
	"github.com/smallsung/gopkg/rpc/jsonrpc"
	"go.uber.org/zap"
)

const (
	initializingState = iota
	runningState
	closedState
)

var (
	ErrGoroutineRunning = fmt.Errorf("main goroutine already running")
	ErrGoroutineClosed  = fmt.Errorf("main goroutine not started")
)

type Goroutine struct {
	config Config
	logger *zap.Logger

	services []Service
	apis     []API

	startStopLock sync.Mutex
	stop          chan error

	state int
	lock  sync.Mutex

	httpServer *httpServer

	inProcRPCServer *rpc.Server
}

func New(config Config) (stack *Goroutine) {
	stack = &Goroutine{
		config: config,
		logger: config.Logger,
	}
	if stack.logger == nil {
		stack.logger = zap.NewNop()
	}

	stack.httpServer = newHttpServer()
	stack.httpServer.logger = stack.logger.Named("http")

	stack.inProcRPCServer = rpc.NewServer(jsonrpc.NewServerCodec)
	stack.inProcRPCServer.Logger = stack.logger.Named("rpc")
	stack.apis = append(stack.apis, stack.builtinAPIs()...)
	return stack
}

func (g *Goroutine) Start() error {
	g.startStopLock.Lock()
	defer g.startStopLock.Unlock()

	g.lock.Lock()
	switch g.state {
	case runningState:
		g.lock.Unlock()
		return ErrGoroutineRunning
	case closedState:
		g.lock.Unlock()
		return ErrGoroutineClosed
	}
	g.state = runningState
	g.lock.Unlock()

	g.logger.Info("starting...")

	if err := g.configuredRPC(); err != nil {
		g.logger.Error("start failure", zap.Error(err))
		return errors.Trace(err)
	}

	if started, err := g.startServices(g.services); err != nil {
		if errs := g.stopServices(started); errs != nil {
			//todo
		}

		g.logger.Error("start failure", zap.Error(err))
		return errors.Trace(err)
	}

	if err := g.openEndpoints(); err != nil {
		g.closeEndpoints()
		if errs := g.stopServices(g.services); errs != nil {
			//todo
		}
		g.logger.Error("start failure", zap.Error(err))
		return errors.Trace(err)
	}

	go g.signal()
	g.logger.Info("start success")
	return nil
}

func (g *Goroutine) Close() (errs []error) {
	g.startStopLock.Lock()
	defer g.startStopLock.Unlock()

	g.lock.Lock()
	defer g.lock.Unlock()

	switch g.state {
	case initializingState:
	case runningState:
		// The service was started, release resources acquired by Start().
		g.logger.Info("shut down...")
		g.state = closedState
		var errs []error
		errs = append(errs, g.stopServices(g.services)...)
		g.closeEndpoints()
		close(g.stop)
	case closedState:
		errs = append(errs, ErrGoroutineClosed)
	}

	return errs
}

func (g *Goroutine) Wait() {
	if g.stop != nil {
		return
	}
	g.stop = make(chan error)
	<-g.stop
}

func (g *Goroutine) signal() {
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sign)

	<-sign
	g.logger.Info("got interrupt, shutting down...")
	go func() {
		if errs := g.Close(); errs != nil {
			for _, err := range errs {
				g.logger.Error("signal", zap.Error(err))
			}
		}
	}()
	for i := 1; ; {
		<-sign
		g.logger.Warn("already shutting down, interrupt more to panic.", zap.Int("times", i))
		i++
	}

}

func (g *Goroutine) openEndpoints() error {
	if g.config.EnableHTTP {
		if err := g.httpServer.setListenerAddr(g.config.HTTPHost, g.config.HTTPPort); err != nil {
			return errors.Annotatef(err, "configured http listener")
		}
		g.logger.Info("starting http")
		if err := g.httpServer.startHTTP(); err != nil {
			return errors.Annotatef(err, "start http")
		}
	}

	return nil
}

func (g *Goroutine) closeEndpoints() {
	if g.config.EnableHTTP {
		g.logger.Info("stopping http")
		g.httpServer.stopHTTP()
		g.logger.Info("stopped http success")
	}

}

func interfaceName(i interface{}) string {
	typ := reflect.ValueOf(i).Elem().Type()
	return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
}
