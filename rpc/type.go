package rpc

import (
	"reflect"
)

type RawMessage = []byte

type (
	MessageID     = RawMessage
	MessageMethod = string
	MessageParams = RawMessage
	MessageResult = RawMessage
	MessageError  struct {
		Code    int64
		Message string
		Data    interface{}
	}
)

type RequestMessage struct {
	ID     MessageID
	Method MessageMethod
	Params MessageParams
}

type ResponseMessage struct {
	ID     MessageID
	Result MessageResult
	Error  *MessageError
}

type RequestMessages struct {
	Batch bool
	Elems []*RequestMessage
}

type ResponseMessages struct {
	Batch bool
	Elems []*ResponseMessage
}

type ServerCodec interface {
	ReadRequest() (RawMessage, error)
	UnmarshalRequest(RawMessage) (*RequestMessages, error)
	UnmarshalRequestParams(MessageParams, Ins) ([]reflect.Value, error)
	MarshalResponseResult(interface{}) (MessageResult, error)
	MarshalResponse(*ResponseMessages) (RawMessage, error)
	WriteResponse(RawMessage) error
	Close() error
}

type ClientCodec interface {
	WriteRequest(RawMessage) error
	MarshalRequestParams(...interface{}) (MessageParams, error)
	MarshalRequest(*RequestMessages) (RawMessage, error)
	UnmarshalResponse(RawMessage) (*ResponseMessages, error)
	UnmarshalResponseResult(MessageResult, interface{}) error
	ReadResponse() (RawMessage, error)
	Close() error
}
