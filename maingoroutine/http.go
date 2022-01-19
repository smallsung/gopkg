package maingoroutine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/smallsung/gopkg/errors"
	"github.com/smallsung/gopkg/rpc"
	"github.com/smallsung/gopkg/rpc/jsonrpc"
	"go.uber.org/zap"
)

type httpRPCHandler struct {
	http.Handler
	rpcServer *rpc.Server
}

var (
	ErrHttpServerRunning = fmt.Errorf("http server already running")
)

type httpServer struct {
	host     string
	port     uint16
	endpoint string

	mutex sync.Mutex

	logger *zap.Logger

	listener     net.Listener
	httpServer   *http.Server
	httpServeMux http.ServeMux

	rpcSupper atomic.Value
}

func newHttpServer() *httpServer {
	s := new(httpServer)
	s.rpcSupper.Store((*httpRPCHandler)(nil))
	return s
}

func (hs *httpServer) setListenerAddr(host string, port uint16) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	if hs.listener != nil || hs.httpServer != nil {
		return errors.Trace(ErrHttpServerRunning)
	}

	hs.host, hs.port = host, port
	hs.endpoint = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return nil
}

func (hs *httpServer) startHTTP() (err error) {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	if hs.listener != nil || hs.httpServer != nil {
		return errors.Trace(ErrHttpServerRunning)
	}
	if hs.endpoint == "" {
		return errors.New("http server not configured")
	}

	var listener net.Listener
	if listener, err = net.Listen("tcp", hs.endpoint); err != nil {
		return errors.Trace(err)
	}
	server := &http.Server{Handler: hs}

	go func() {
		hs.logger.Info("http server running", zap.String("addr", hs.listener.Addr().String()))
		_ = server.Serve(listener)
	}()

	hs.listener, hs.httpServer = listener, server
	return nil
}

func (hs *httpServer) stopHTTP() {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()
	if hs.httpServer != nil {
		if err := hs.httpServer.Shutdown(context.Background()); err != nil {
			hs.logger.Warn("http server shutdown", zap.Error(err))
		}
	}
	hs.host, hs.port, hs.endpoint = "", 0, ""
	hs.httpServer, hs.listener = nil, nil
}

func (hs *httpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	if handler, pattern := hs.httpServeMux.Handler(request); pattern != "" {
		handler.ServeHTTP(writer, request)
		return
	}

	if rpcHandler := hs.rpcSupper.Load().(*httpRPCHandler); rpcHandler != nil {
		rpcHandler.ServeHTTP(writer, request)
		return
	}

	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprint(writer, http.StatusText(http.StatusInternalServerError))
}

func (hs *httpServer) enableRPC(apis []API) error {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()

	if hs.rpcSupper.Load().(*httpRPCHandler) != nil {
		return errors.New("JSON-RPC over HTTP is already enabled")
	}
	rpcServer := rpc.NewServer(jsonrpc.NewServerCodec)
	rpcServer.Logger = hs.logger.Named("rpc")

	for _, api := range apis {
		if api.Public {
			if err := rpcServer.Register(api.Namespace, api.Service); err != nil {
				hs.logger.Warn("register http rpc failure", zap.String("n", api.Namespace))
				return errors.Trace(err)
			}
		}
	}

	hs.rpcSupper.Store(&httpRPCHandler{
		rpcServer: rpcServer,
		Handler:   wrapJsonRPCHttpHandler(rpcServer),
	})

	return nil
}

func (hs *httpServer) disableRPC() {
	if rpcHandler := hs.rpcSupper.Load().(*httpRPCHandler); rpcHandler != nil {
		hs.rpcSupper.Store((*httpRPCHandler)(nil))
		rpcHandler.rpcServer.Shutdown()
	}
}

func wrapJsonRPCHttpHandler(rpcHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		jsonrpc.HttpHandlers.ValidHeader.ServeHTTP(response, request)
		rpcHandler.ServeHTTP(response, request)
	})
}
