package rpc

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"

	"go.uber.org/zap"
)

type NewServerCodecFunc func(io.ReadWriteCloser) ServerCodec

type Server struct {
	running uint32

	Logger *zap.Logger

	newCodec NewServerCodecFunc

	codecCount uint64
	codecs     map[uint64]ServerCodec

	registry registry
}

func NewServer(newCodecFunc NewServerCodecFunc) *Server {
	s := &Server{
		running:  1,
		Logger:   zap.NewNop(),
		newCodec: newCodecFunc,
		codecs:   make(map[uint64]ServerCodec),
	}
	_ = s.Register(builtinServiceName, builtinService{s})
	return s
}

func (s *Server) Register(namespaces string, receiver interface{}) error {
	return s.registry.register(namespaces, receiver)
}

func (s *Server) Accept(listener net.Listener) {
	ctx := context.Background()
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.Logger.Error("rpc.Server.Accept:", zap.Error(err))
			return
		}
		go s.ServeConn(ctx, conn)
	}
}

func (s *Server) ServeConn(ctx context.Context, conn io.ReadWriteCloser) {
	codec := s.newCodec(conn)
	s.ServeCodec(ctx, codec)
}

func (s *Server) ServeCodec(ctx context.Context, codec ServerCodec) {
	defer codec.Close()
	if atomic.LoadUint32(&s.running) == 0 {
		return
	}
	s.codecs[atomic.AddUint64(&s.codecCount, 1)] = codec

	baseCtx := ctx
	for {
		var err error
		var raw RawMessage
		if raw, err = s.readRequest(codec); err != nil {
			s.Logger.Warn("codec.ReadRequest", zap.Error(err))
			continue
		}
		ctx = context.WithValue(baseCtx, "", raw)
		go s.serveRequest(ctx, codec, raw)
	}

}

//ServeRequest 处理单次，不会关闭连接
func (s *Server) ServeRequest(ctx context.Context, codec ServerCodec) (err error) {
	if atomic.LoadUint32(&s.running) == 0 {
		return
	}

	var raw []byte
	if raw, err = s.readRequest(codec); err != nil {
		s.Logger.Warn("codec.ReadRequest", zap.Error(err))
		return err
	}
	ctx = context.WithValue(ctx, "", raw)
	return s.serveRequest(ctx, codec, raw)
}

func (s *Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	select {
	case <-ctx.Done():
		return
	default:
	}

	conn := &httpServerConn{response: response, request: request}
	codec := s.newCodec(conn)
	defer codec.Close()
	if err := s.ServeRequest(request.Context(), codec); err != nil {
		s.Logger.Warn("server.ServeHTTP", zap.Error(err))
	}
}

func (s *Server) Shutdown() {
	if atomic.CompareAndSwapUint32(&s.running, 1, 0) {
		s.Logger.Info("rpc server shutting down")
		for _, codec := range s.codecs {
			_ = codec.Close()
		}
	}
}

func (s *Server) readRequest(codec ServerCodec) (raw []byte, err error) {
	if raw, err = codec.ReadRequest(); err != nil {
		_ = s.writeErrorResponse(codec, ErrParseError)
		return nil, err
	}
	s.Logger.Debug("server.readRequest", zap.String("raw", string(raw)))
	return raw, nil
}

func (s *Server) unmarshalRequest(codec ServerCodec, raw []byte) (*RequestMessages, error) {
	if requests, err := codec.UnmarshalRequest(raw); err == nil {
		return requests, nil
	} else {
		s.Logger.Warn("codec.UnmarshalRequest", zap.Error(err))
		return nil, err
	}
}

func (s *Server) writeErrorResponse(codec ServerCodec, err error) error {
	return s.writeResponse(codec, NewResponseMessages(errorResponseMessage(err)))
}

func (s *Server) writeResponse(codec ServerCodec, responses *ResponseMessages) (err error) {
	var raw []byte
	if raw, err = s.marshalResponse(codec, responses); err == nil {
		if err = codec.WriteResponse(raw); err != nil {
			s.Logger.Debug("codec.WriteResponse", zap.Error(err))
			return err
		}
		return nil
	}
	// 对于编码错误的情况，向客户端反馈。  ErrInternalError
	// 对于确定的结构体，确定的值，这里不应该会出现错误。
	if raw, err := s.marshalResponse(codec, NewResponseMessages(errorResponseMessage(ErrInternalError))); err == nil {
		if err := codec.WriteResponse(raw); err != nil {
			s.Logger.Debug("codec.WriteResponse", zap.Error(err))
		}
	}

	return err
}

func (s *Server) marshalResponse(codec ServerCodec, responses *ResponseMessages) ([]byte, error) {
	if raw, err := codec.MarshalResponse(responses); err != nil {
		s.Logger.Warn("codec.MarshalResponse", zap.Error(err))
		return nil, err
	} else {
		return raw, nil
	}
}

func (s *Server) serveRequest(ctx context.Context, codec ServerCodec, raw []byte) error {
	requests, err := s.unmarshalRequest(codec, raw)
	//包含空数组的rpc调用:
	//--> []
	//<-- {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
	if err != nil || len(requests.Elems) == 0 {
		_ = s.writeErrorResponse(codec, ErrInvalidRequest)
		return ErrInvalidRequest
	}

	ctx = context.WithValue(ctx, "", requests)
	h := &handler{
		registry: &s.registry,
		codec:    codec,
	}
	responses := h.handleMessages(ctx, requests)

	if len(responses.Elems) == 0 {
		return nil
	}

	return s.writeResponse(codec, responses)
}
