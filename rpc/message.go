package rpc

import (
	"fmt"

	"github.com/smallsung/gopkg/errors"
)

func (rm *RequestMessage) ResponseResult(result MessageResult) *ResponseMessage {
	return &ResponseMessage{ID: rm.ID, Result: result}
}

func (rm *RequestMessage) ResponseError(err error) *ResponseMessage {
	response := errorResponseMessage(err)
	response.ID = rm.ID
	return response
}

func (rm *RequestMessage) IsNotification() bool {
	return rm.ID == nil && rm.Method != ""
}

func (rm *RequestMessage) IsCallBack() bool {
	return rm.HasValidID() && rm.Method != ""
}

func (rm *RequestMessage) HasValidID() bool {
	return len(rm.ID) > 0 && rm.ID[0] != '{' && rm.ID[0] != '['
}

func (rm *ResponseMessage) IsResponse() bool {
	return rm.HasValidID() && ((rm.Result == nil && rm.Error != nil) || (rm.Result != nil && rm.Error == nil))
}
func (rm *ResponseMessage) HasValidID() bool {
	return len(rm.ID) > 0 && rm.ID[0] != '{' && rm.ID[0] != '['
}

var debugErrorOnClient bool

func EnableDebugErrorOnClient() {
	debugErrorOnClient = true
}

const defaultErrorCode = -32000

func errorResponseMessage(err error) *ResponseMessage {
	me := &MessageError{Code: defaultErrorCode}
	if debugErrorOnClient {
		me.Message = fmt.Sprintf("%+v", err)
	} else {
		var em ErrorMessage
		if errors.As(err, &em) {
			me.Message = em.RPCErrorMessage()
		} else {
			me.Message = err.Error()
		}
	}

	var ec ErrorCode
	if errors.As(err, &ec) {
		me.Code = ec.RPCErrorCode()
	}
	var ed ErrorData
	if errors.As(err, &ed) {
		me.Data = ed.RPCErrorData()
	}
	return &ResponseMessage{ID: nil, Error: me}
}

func NewRequestMessages(messages ...*RequestMessage) *RequestMessages {
	rms := new(RequestMessages)
	for _, m := range messages {
		rms.Append(m)
	}
	return rms
}

func (rms *RequestMessages) Append(m *RequestMessage) {
	rms.Elems = append(rms.Elems, m)
	if len(rms.Elems) > 1 {
		rms.Batch = true
	}
}

func NewResponseMessages(messages ...*ResponseMessage) *ResponseMessages {
	rms := new(ResponseMessages)
	for _, m := range messages {
		rms.Append(m)
	}
	return rms
}

func (rms *ResponseMessages) Append(message *ResponseMessage) {
	rms.Elems = append(rms.Elems, message)
	if len(rms.Elems) > 1 {
		rms.Batch = true
	}
}
