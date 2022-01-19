package rpc

import (
	"fmt"
)

var ErrCallbackNameExist = fmt.Errorf("callback name exist")

type (
	ErrorMessage interface{ RPCErrorMessage() string }
	ErrorCode    interface{ RPCErrorCode() int64 }
	ErrorData    interface{ RPCErrorData() interface{} }
)

func (err *MessageError) Error() string { return err.Message }

//func (err messageError) RPCErrorMessage() string   { return err.Message }
//func (err messageError) RPCErrorCode() int64         { return err.Code }
//func (err messageError) RPCErrorData() interface{} { return err.Data }

type preDefinedError struct {
	code    int64
	message string
}

func (err *preDefinedError) Error() string       { return err.message }
func (err *preDefinedError) RPCErrorCode() int64 { return err.code }

// 下面定义的错误类型是JSON-RPC内置错误
var (
	//ErrInvalidRequest The JSON sent is not a valid Request object.
	ErrInvalidRequest = &preDefinedError{code: -32600, message: "Invalid Request"}
	//ErrMethodNotFound The method does not exist / is not available.
	ErrMethodNotFound = &preDefinedError{code: -32601, message: "Method not found"}
	//ErrInvalidParams Invalid method parameter(s).
	ErrInvalidParams = &preDefinedError{code: -32602, message: "Invalid params"}
	//ErrInternalError Internal JSON-RPC error.
	ErrInternalError = &preDefinedError{code: -32603, message: "Internal error"}
	//ErrParseError Invalid JSON was received by the server.An error occurred on the server while parsing the JSON text.
	ErrParseError = &preDefinedError{code: -32700, message: "Parse error"}
)
