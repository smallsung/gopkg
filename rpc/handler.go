package rpc

import (
	"context"
	"reflect"
	"sync"
)

type handler struct {
	codec    ServerCodec
	registry *registry
}

func (h *handler) handleMessages(ctx context.Context, requests *RequestMessages) *ResponseMessages {
	responses := NewResponseMessages()
	responses.Batch = requests.Batch
	wg := sync.WaitGroup{}
	for _, r := range requests.Elems {
		wg.Add(1)
		request := r
		go func() {
			defer wg.Done()
			if response := h.handleMessage(ctx, request); response != nil {
				responses.Append(response)
			}

		}()
	}
	wg.Wait()
	return responses
}

func (h *handler) handleMessage(ctx context.Context, request *RequestMessage) *ResponseMessage {
	switch {
	case request.IsNotification():
		return h.handleNotification(ctx, request)
	case request.IsCallBack():
		return h.handleCallBack(ctx, request)
	case request.HasValidID():
		return request.ResponseError(ErrInvalidRequest)
	default:
		return errorResponseMessage(ErrInvalidRequest)
	}
}

func (h *handler) handleNotification(ctx context.Context, request *RequestMessage) *ResponseMessage {
	return nil
}

func (h *handler) handleCallBack(ctx context.Context, request *RequestMessage) *ResponseMessage {
	var cb *callback
	if cb = h.registry.callback(request.Method); cb == nil {
		return request.ResponseError(ErrMethodNotFound)
	}

	var err error
	var arguments []reflect.Value
	if arguments, err = h.codec.UnmarshalRequestParams(request.Params, cb.ins); err != nil {
		return request.ResponseError(ErrInvalidParams)
	}

	var result interface{}
	if result, err = cb.call(ctx, arguments); err != nil {
		return request.ResponseError(err)

	}

	var raw MessageResult
	if raw, err = h.codec.MarshalResponseResult(result); err != nil {
		return request.ResponseError(ErrInternalError)
	}

	return request.ResponseResult(raw)
}
