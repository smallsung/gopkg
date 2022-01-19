package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"reflect"

	"github.com/smallsung/gopkg/rpc"
)

type serverCodec struct {
	rwc     io.ReadWriteCloser
	decoder *json.Decoder
}

func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{rwc: conn, decoder: json.NewDecoder(conn)}
}

func (codec *serverCodec) ReadRequest() (rpc.RawMessage, error) {
	var rawMessage json.RawMessage
	if err := codec.decoder.Decode(&rawMessage); err != nil {
		return nil, err
	}
	return rawMessage, nil
}

func (codec *serverCodec) UnmarshalRequest(rawMessage rpc.RawMessage) (*rpc.RequestMessages, error) {
	if requests, err := parseRequestRawMessage(rawMessage); err != nil {
		return nil, err
	} else {
		return requests, nil
	}
}

func (codec *serverCodec) UnmarshalRequestParams(params rpc.MessageParams, ins rpc.Ins) ([]reflect.Value, error) {
	//not Supported by-name
	if !isArrayRawMessage(params) {
		return nil, rpc.ErrInvalidParams
	}
	return parsePositionalArguments(params, ins.Position)
}

func (codec *serverCodec) MarshalResponseResult(i interface{}) (rpc.MessageResult, error) {
	return json.Marshal(i)
}

func (codec *serverCodec) MarshalResponse(responses *rpc.ResponseMessages) (rpc.RawMessage, error) {
	s := make([]*ResponseMessage, 0, len(responses.Elems))
	for _, m := range responses.Elems {
		if m.ID == nil {
			m.ID = null
		}
		s = append(s, toResponseMessage(m))
	}
	if responses.Batch {
		return json.Marshal(s)
	} else {
		return json.Marshal(s[0])
	}
}

func (codec *serverCodec) WriteResponse(rawMessage rpc.RawMessage) error {
	_, err := codec.rwc.Write(rawMessage)
	return err
}

func (codec *serverCodec) Close() error {
	return codec.rwc.Close()
}

const (
	contentType             = "application/json"
	maxRequestContentLength = 1024 * 1024 * 5
)

var (
	acceptedContentTypes = []string{contentType}
)

var HttpHandlers = struct {
	ValidHeader http.Handler
}{
	ValidHeader: http.HandlerFunc(validateRequestHeader),
}

func validateRequestHeader(response http.ResponseWriter, request *http.Request) {
	cancel := true
	ctx, cancelFunc := context.WithCancel(request.Context())
	defer func() {
		if cancel {
			cancelFunc()
		}
	}()
	*request = *request.WithContext(ctx)

	if request.Method == http.MethodGet && request.ContentLength == 0 && request.URL.RawQuery == "" {
		response.WriteHeader(http.StatusOK)
		return
	}

	if request.Method != http.MethodPost {
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if request.ContentLength > maxRequestContentLength {
		http.Error(response, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("content-type"))
	if err != nil {
		http.Error(response, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}
	var flag bool
	for _, accepted := range acceptedContentTypes {
		if accepted == mediaType {
			flag = true
		}
	}
	if !flag {
		http.Error(response, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	response.Header().Set("content-type", contentType)
	cancel = false
}
