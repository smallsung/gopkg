package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"

	"github.com/smallsung/gopkg/errors"
	"github.com/smallsung/gopkg/rpc"
)

const defaultJsonRpcVersion = "2.0"

type MessageError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RequestMessage struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseMessage struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Version string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MessageError   `json:"error,omitempty"`
}

var null = json.RawMessage("null")

func toRPCRequestMessage(request *RequestMessage) *rpc.RequestMessage {
	return &rpc.RequestMessage{
		ID:     request.ID,
		Method: request.Method,
		Params: request.Params,
	}
}

func toRequestMessage(request *rpc.RequestMessage) *RequestMessage {
	return &RequestMessage{
		ID:      request.ID,
		Version: defaultJsonRpcVersion,
		Method:  request.Method,
		Params:  request.Params,
	}
}

func toResponseMessage(response *rpc.ResponseMessage) *ResponseMessage {
	to := &ResponseMessage{
		ID:      response.ID,
		Version: defaultJsonRpcVersion,
		Result:  response.Result,
	}
	if response.Error != nil {
		to.Error = &MessageError{
			Code:    response.Error.Code,
			Message: response.Error.Message,
			Data:    response.Error.Data,
		}
	}
	return to
}

func toRPCResponseMessage(response *ResponseMessage) *rpc.ResponseMessage {
	to := &rpc.ResponseMessage{
		ID:     response.ID,
		Result: response.Result,
	}
	if response.Error != nil {
		to.Error = &rpc.MessageError{
			Code:    response.Error.Code,
			Message: response.Error.Message,
			Data:    response.Error.Data,
		}
	}
	return to
}

//func rpcMessage2Message(rpcMessage *rpc.Message) *Message {
//	message := &Message{
//		ID:      rpcMessage.ID,
//		Version: defaultJsonRpcVersion,
//		Method:  rpcMessage.Method,
//		Params:  rpcMessage.Params,
//		Result:  rpcMessage.Result,
//		Error:   nil,
//	}
//	if rpcMessage.Error != nil {
//		message.Error = &MessageError{
//			Code:    rpcMessage.Error.Code,
//			Message: rpcMessage.Error.Message,
//			Data:    rpcMessage.Error.Data,
//		}
//	}
//	return message
//}
//
//func message2RpcMessage(message *Message) *rpc.Message {
//	rpcMessage := &rpc.Message{
//		ID:     message.ID,
//		Method: message.Method,
//		Params: message.Params,
//		Result: message.Result,
//		Error:  nil,
//	}
//	if message.Error != nil {
//		rpcMessage.Error = &rpc.MessageError{
//			Code:    message.Error.Code,
//			Message: message.Error.Message,
//			Data:    message.Error.Data,
//		}
//	}
//	return rpcMessage
//}

//func errorResponseMessage(err error) *Message {
//	mre := &rpc.messageError{Message: err.Error(), Code: defaultMessageErrorCode}
//	if ec, ok := err.(rpc.RPCErrorCode); ok {
//		mre.Code = ec.RPCErrorCode()
//	}
//	if ed, ok := err.(rpc.RPCErrorData); ok {
//		mre.Data = ed.RPCErrorData()
//	}
//
//	return &Message{ID: nil, Version: defaultJsonRpcVersion, Error: mre}
//}

//parseRawMessage
//https://www.jsonrpc.org/specification#examples
//包含无效请求对象的rpc调用:
//	--> {"rpc": "2.0", "method": 1, "params": "bar"}
//	<-- {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//非空且无效的rpc批量调用:
//	--> [1]
//	<-- [
//	    {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//	    ]
//无效的rpc批量调用:
//	--> [1,2,3]
//	<-- [
//	    {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//	    {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//	    {"rpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//	    ]
func parseRequestRawMessage(rawMessage json.RawMessage) (*rpc.RequestMessages, error) {
	requests := rpc.NewRequestMessages()
	if !isArrayRawMessage(rawMessage) {
		var v RequestMessage
		if err := json.Unmarshal(rawMessage, &v); err != nil {
			return nil, err
		}
		if v.Version != defaultJsonRpcVersion {
			return nil, rpc.ErrInvalidRequest
		}
		requests.Append(toRPCRequestMessage(&v))
	} else {
		decoder := json.NewDecoder(bytes.NewReader(rawMessage))
		_, _ = decoder.Token() // skip '['
		for decoder.More() {
			var v RequestMessage
			if err := decoder.Decode(&v); err != nil {
				requests.Append(new(rpc.RequestMessage))
				continue
			}
			if v.Version != defaultJsonRpcVersion {
				requests.Append(new(rpc.RequestMessage))
				continue
			}
			requests.Append(toRPCRequestMessage(&v))
		}
		requests.Batch = true
	}

	for i, message := range requests.Elems {
		if string(message.ID) == string(null) {
			requests.Elems[i].ID = nil
		}
	}

	return requests, nil
}

func parseResponseRawMessage(rawMessage json.RawMessage) (*rpc.ResponseMessages, error) {
	responses := rpc.NewResponseMessages()
	if !isArrayRawMessage(rawMessage) {
		var v ResponseMessage
		if err := json.Unmarshal(rawMessage, &v); err != nil {
			return nil, err
		}
		if v.Version != defaultJsonRpcVersion {
			return nil, rpc.ErrInvalidRequest
		}
		responses.Append(toRPCResponseMessage(&v))
	} else {
		decoder := json.NewDecoder(bytes.NewReader(rawMessage))
		_, _ = decoder.Token() // skip '['
		for decoder.More() {
			var v ResponseMessage
			if err := decoder.Decode(&v); err != nil {
				responses.Append(new(rpc.ResponseMessage))
				continue
			}
			if v.Version != defaultJsonRpcVersion {
				responses.Append(new(rpc.ResponseMessage))
				continue
			}
			responses.Append(toRPCResponseMessage(&v))
		}
		responses.Batch = true
	}

	for i, message := range responses.Elems {
		if string(message.ID) == string(null) {
			responses.Elems[i].ID = nil
		}
	}

	return responses, nil
}

// isArrayRawMessage 当第一个非空白字符是 [ 时 rawMessage 是批处理消息
func isArrayRawMessage(bytes []byte) bool {
	for _, c := range bytes {
		// skip insignificant whitespace (http://www.ietf.org/rfc/rfc4627.txt)
		if c == 0x20 || c == 0x09 || c == 0x0a || c == 0x0d {
			continue
		}
		return c == '['
	}
	return false
}

func parsePositionalArguments(params rpc.MessageParams, types []reflect.Type) ([]reflect.Value, error) {
	decoder := json.NewDecoder(bytes.NewReader(params))
	var args []reflect.Value
	token, err := decoder.Token()
	switch {
	case err == io.EOF:
	case err != nil:
		return nil, errors.Trace(err)
	case token == json.Delim('['):
		if args, err = parseArgumentArray(decoder, types); err != nil {
			return nil, errors.Trace(err)
		}
	default:
		return nil, errors.New("params MUST be an Array, containing the values in the Server expected order.")
	}
	//Set any missing args to nil.
	for i := len(args); i < len(types); i++ {
		// todo go-ethereum 中这个地方对非指针变量有限制，还不明白这么做的意义
		switch types[i].Kind() {
		case reflect.Ptr, reflect.Interface:
		default:
			return nil, errors.Format("missing value for required argument %d", i)
		}
		args = append(args, reflect.Zero(types[i]))
	}
	return args, nil
}

func parseArgumentArray(decoder *json.Decoder, types []reflect.Type) ([]reflect.Value, error) {
	quantity := len(types)
	args := make([]reflect.Value, 0, quantity)
	for i := 0; decoder.More(); i++ {
		if i >= quantity {
			return nil, errors.Format("too many arguments, want at most %d", len(types))
		}
		val := reflect.New(types[i])
		if err := decoder.Decode(val.Interface()); err != nil {
			return nil, errors.Trace(err)
		}
		//if val.IsNil() && val.Kind() != reflect.Ptr {
		//	return args, errors.Errorf("missing value for required argument %d", i)
		//}
		args = append(args, val.Elem())
	}
	// Read end of args array.
	if _, err := decoder.Token(); err != nil {
		return nil, errors.Trace(err)
	}
	return args, nil
}
