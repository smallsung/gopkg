package rpc

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/smallsung/gopkg/errors"
	"go.uber.org/zap"
)

type BatchElem struct {
	Method MessageMethod
	Params []interface{}
	Result interface{}
	Error  error

	id   MessageID
	done bool
}

type Call struct {
	requests    *RequestMessages
	requestsRaw RawMessage
	waitGroup   sync.WaitGroup

	Result interface{}
	Error  error

	Elems []BatchElem
	elems map[string]*BatchElem

	Done chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

func (c *Call) wait() {
	c.waitGroup.Wait()
	if c.Done != nil {
		c.done()
	}
}

type Client struct {
	Logger *zap.Logger

	isHttp bool

	idCounter uint64

	codec ClientCodec

	calls map[string]*Call

	readError chan error
	readChan  chan *ResponseMessages
	sendChan  chan *Call
	sentChan  chan error

	closeChan chan error
}

func (c *Client) CallAsync(ctx context.Context, done chan *Call, method MessageMethod, result interface{}, params ...interface{}) *Call {
	if cap(done) == 0 {
		panic("rpc: done channel is unbuffered")
	}
	call := &Call{
		requests: NewRequestMessages(),
		Result:   result,
		Done:     done,
	}

	var err error
	var request *RequestMessage
	if request, err = c.buildMessage(method, params...); err != nil {
		call.Error = errors.Trace(err)
		call.done()
		return call
	}
	call.requests.Append(request)

	if err = c.sendCall(ctx, call); err != nil {
		call.Error = errors.Annotate(err, "sendCall")
		call.done()
		return call
	}

	return call
}

func (c *Client) Call(ctx context.Context, method MessageMethod, result interface{}, params ...interface{}) (err error) {
	call := <-c.CallAsync(ctx, make(chan *Call, 1), method, result, params...).Done
	return call.Error
}

func (c *Client) BathAsync(ctx context.Context, done chan *Call, elems ...BatchElem) *Call {
	if cap(done) == 0 {
		panic("rpc: done channel is unbuffered")
	}
	call := &Call{
		requests: NewRequestMessages(),
		Done:     done,
		Elems:    elems,
		elems:    make(map[string]*BatchElem),
	}

	l := len(elems)
	if l == 0 {
		call.Error = ErrInvalidParams
		call.done()
		return call
	}
	call.requests.Batch = true

	var err error
	for i := 0; i < l; i++ {
		var request *RequestMessage
		if request, err = c.buildMessage(elems[i].Method, elems[i].Params...); err != nil {
			err = errors.Trace(err)
			elems[i].Error = err
			break
		}
		elem := &elems[i]
		elem.id = request.ID
		call.elems[string(request.ID)] = elem
		call.requests.Append(request)
	}
	if err != nil {
		call.Error = err
		call.done()
		return call
	}

	if err := c.sendCall(ctx, call); err != nil {
		call.Error = errors.Annotate(err, "sendCall")
		call.done()
		return call
	}
	return call
}

func (c *Client) Bath(ctx context.Context, elems ...BatchElem) error {
	call := <-c.BathAsync(ctx, make(chan *Call, 1), elems...).Done
	return call.Error
}

func (c *Client) Notice(ctx context.Context, method string, params ...interface{}) (err error) {
	call := &Call{
		requests: new(RequestMessages),
	}
	var request *RequestMessage
	if request, err = c.buildMessage(method, params...); err != nil {
		return errors.Annotate(err, "buildMessage")
	}
	request.ID = nil
	call.requests.Append(request)
	if err = c.sendCall(ctx, call); err != nil {
		return errors.Annotate(err, "sendCall")
	}
	return nil
}

func (c *Client) buildMessage(method string, params ...interface{}) (*RequestMessage, error) {
	message := new(RequestMessage)
	if len(params) > 0 {
		var err error
		if message.Params, err = c.codec.MarshalRequestParams(params...); err != nil {
			return nil, errors.Annotate(err, "codec.MarshalRequestParams")
		}
	}
	message.ID, message.Method = c.nextID(), method
	return message, nil
}

func (c *Client) nextID() MessageID {
	id := atomic.AddUint64(&c.idCounter, 1)
	return strconv.AppendUint(nil, id, 10)
}

func (c *Client) addCall(call *Call) {
	for _, request := range call.requests.Elems {
		if !request.IsNotification() {
			c.calls[string(request.ID)] = call
			call.waitGroup.Add(1)
		}
	}
}

func (c *Client) removeCall(call *Call) {
	for _, request := range call.requests.Elems {
		if !request.IsNotification() {
			delete(c.calls, string(request.ID))
			call.waitGroup.Done()
		}
	}
}

func (c *Client) handleMessages(responses *ResponseMessages) {
	for _, response := range responses.Elems {
		switch {
		case response.IsResponse():
			c.handleResponse(response)
		case response.HasValidID():
			//todo
		default:
			//todo
		}
	}
}

func (c *Client) handleResponse(response *ResponseMessage) {
	id := string(response.ID)
	call := c.calls[id]
	switch {
	case call == nil:
		c.Logger.Warn("client.handleMessages:unsolicited RPC response", zap.String("id", id))

	case response.Error != nil:
		if call.requests.Batch {
			call.elems[id].Error = response.Error
		}
		call.Error = response.Error
		call.waitGroup.Done()
	default:
		var result interface{}
		if !call.requests.Batch {
			result = call.Result
		} else {
			result = call.elems[id].Result
		}

		if result != nil {
			err := errors.Annotate(c.codec.UnmarshalResponseResult(response.Result, result), "codec.UnmarshalResponseResult")
			if err != nil {
				if call.requests.Batch {
					call.elems[id].Error = err
				}
				call.Error = err
			}
		}

		call.waitGroup.Done()
	}
	delete(c.calls, id)
}

func (c *Client) loop() {
	var (
		sendChan = c.sendChan
		lastCall *Call
	)

	go c.read()
	for {
		select {

		case <-c.closeChan:
			return

		case call := <-sendChan:
			// 暂停其他请求
			sendChan = nil
			lastCall = call
			c.addCall(lastCall)

		case err := <-c.sentChan:
			if err != nil {
				c.removeCall(lastCall)
			}
			sendChan, lastCall = c.sendChan, nil

		case err := <-c.readError:
			if err != nil {
				c.Logger.Warn("client.readError", zap.Error(err))
			}

		case responses := <-c.readChan:
			c.handleMessages(responses)
		}
	}
}

func (c *Client) sendCall(ctx context.Context, call *Call) (err error) {
	if call.requestsRaw, err = c.codec.MarshalRequest(call.requests); err != nil {
		return errors.Annotate(err, "codec.MarshalRequest")
	}

	if c.isHttp {
		return errors.Annotate(c.sendHttp(ctx, call), "client.sendCall")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.sendChan <- call:
		err = c.writeRequest(call.requestsRaw)
		c.sentChan <- err
		if err == nil {
			go call.wait()
		}
		return errors.Annotate(err, "client.sendCall")
	}
}

func (c *Client) sendHttp(ctx context.Context, call *Call) (err error) {
	httpCodec := c.codec.(*httpClientCodec)
	var request *http.Request
	if request, err = http.NewRequestWithContext(ctx, http.MethodPost, httpCodec.url.String(), bytes.NewReader(call.requestsRaw)); err != nil {
		return errors.Trace(err)
	}

	go func() {
		var err error
		defer func() {
			if err != nil {
				call.Error = err
			}
			call.done()
		}()

		var httpResponse *http.Response
		if httpResponse, err = httpCodec.client.Do(request); err != nil {
			err = errors.Trace(err)
			return
		}

		if httpResponse.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(httpResponse.Body)
			err = newHttpError(httpResponse.Status, body)
			return
		}
		defer httpResponse.Body.Close()

		var rawMessage RawMessage
		if rawMessage, err = ioutil.ReadAll(httpResponse.Body); err != nil {
			err = errors.Trace(err)
			return
		}
		c.Logger.Debug("client.readResponse", zap.String("raw", string(rawMessage)))

		var responses *ResponseMessages
		if responses, err = c.codec.UnmarshalResponse(rawMessage); err != nil {
			err = errors.Annotate(err, "httpCodec.UnmarshalResponse")
			return
		}

		if !call.requests.Batch {
			switch {
			case call.requests.Elems[0].IsNotification():
			case responses.Elems[0].Error != nil:
				err = errors.Trace(err)
			default:
				if err = c.codec.UnmarshalResponseResult(responses.Elems[0].Result, call.Result); err != nil {
					err = errors.Annotate(err, "httpCodec.UnmarshalResponseResult")
				}
			}
			return
		}

		for _, response := range responses.Elems {
			switch {
			case response.Error != nil:
				call.elems[string(response.ID)].Error = response.Error
			default:
				if err = c.codec.UnmarshalResponseResult(response.Result, call.elems[string(response.ID)].Result); err != nil {
					call.elems[string(response.ID)].Error = errors.Annotate(err, "httpCodec.UnmarshalResponseResult")
				}
			}
			call.elems[string(response.ID)].done = true
		}

		for _, elem := range call.elems {
			if elem.done {
				continue
			}
			elem.Error = io.EOF
		}
		return
	}()

	return nil
}

func (c *Client) writeRequest(raw RawMessage) (err error) {
	if err = c.codec.WriteRequest(raw); err != nil {
		return errors.Annotate(err, "codec.WriteRequest")
	}
	return nil
}

func (c *Client) read() {
	for {
		var raw RawMessage
		var err error
		if raw, err = c.codec.ReadResponse(); err != nil {
			c.readError <- errors.Annotate(err, "codec.ReadResponse")
			continue
		}

		c.Logger.Debug("client.readResponse", zap.String("raw", string(raw)))

		var response *ResponseMessages
		if response, err = c.codec.UnmarshalResponse(raw); err != nil {
			c.readError <- errors.Annotate(err, "codec.UnmarshalResponse")
			continue
		}
		c.readChan <- response
	}
}

func (c *Client) Close() {
	c.closeChan <- nil
}

func NewClient(codec ClientCodec) *Client {
	_, isHttp := codec.(*httpClientCodec)
	c := &Client{
		Logger:    zap.NewNop(),
		isHttp:    isHttp,
		idCounter: 0,
		codec:     codec,
		calls:     make(map[string]*Call),
		readError: make(chan error),
		readChan:  make(chan *ResponseMessages),
		sendChan:  make(chan *Call),
		sentChan:  make(chan error),
		closeChan: make(chan error),
	}
	if !isHttp {
		go c.loop()
	}
	return c
}

type NewClientCodecFunc func(io.ReadWriteCloser) ClientCodec

func Dial(ctx context.Context, endpoint string, newCodecFunc NewClientCodecFunc) (client *Client, err error) {
	var URL *url.URL
	if URL, err = url.Parse(endpoint); err != nil {
		return nil, errors.Trace(err)
	}
	switch URL.Scheme {
	case "http", "https":
		return DialHTTP(ctx, URL, newCodecFunc), nil
	//case "ws", "wss":
	//	return DialWebsocket(ctx, rawurl, "")
	//case "stdio":
	//	return DialStdIO(ctx)
	case "":
		return DialIPC(ctx, endpoint, newCodecFunc)
	default:
		return nil, errors.Format("no known transport for URL scheme %q", URL.Scheme)
	}
}

func DialInProc(ctx context.Context, endpoint *Server, newCodecFunc NewClientCodecFunc) *Client {
	p1, p2 := net.Pipe()
	go endpoint.ServeConn(context.Background(), p1)
	codec := newCodecFunc(p2)
	return NewClient(codec)
}
