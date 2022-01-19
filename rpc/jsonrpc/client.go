package jsonrpc

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/smallsung/gopkg/rpc"
)

type clientCodec struct {
	rwc     io.ReadWriteCloser
	decoder *json.Decoder
}

func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{rwc: conn, decoder: json.NewDecoder(conn)}
}

func (codec *clientCodec) WriteRequest(rawMessage rpc.RawMessage) error {
	_, err := codec.rwc.Write(rawMessage)
	return err
}

func (codec *clientCodec) MarshalRequest(requests *rpc.RequestMessages) (rpc.RawMessage, error) {
	s := make([]*RequestMessage, 0, len(requests.Elems))
	for _, m := range requests.Elems {
		s = append(s, toRequestMessage(m))
	}
	if requests.Batch {
		return json.Marshal(s)
	} else {
		return json.Marshal(s[0])
	}
}

func (codec *clientCodec) MarshalRequestParams(i ...interface{}) (rpc.MessageParams, error) {
	return json.Marshal(i)
}

func (codec *clientCodec) UnmarshalResponse(rawMessage rpc.RawMessage) (*rpc.ResponseMessages, error) {
	if responses, err := parseResponseRawMessage(rawMessage); err != nil {
		return nil, err
	} else {
		return responses, nil
	}
}

func (codec *clientCodec) UnmarshalResponseResult(rawMessage rpc.MessageResult, i interface{}) error {
	return json.Unmarshal(rawMessage, i)
}

func (codec *clientCodec) ReadResponse() (rpc.RawMessage, error) {
	var rawMessage json.RawMessage
	if err := codec.decoder.Decode(&rawMessage); err != nil {
		return nil, err
	}
	return rawMessage, nil
}

func (codec *clientCodec) Close() error {
	return codec.rwc.Close()
}

type httpRoundTripperFunc func(request *http.Request) (*http.Response, error)

func (f httpRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

var HttpRoundTripper = struct {
	ValidHeader http.RoundTripper
}{
	ValidHeader: httpRoundTripperFunc(validateResponseHeader),
}

func validateResponseHeader(request *http.Request) (*http.Response, error) {
	request.Header.Set("accept", contentType)
	request.Header.Set("content-type", contentType)
	if response, err := http.DefaultTransport.RoundTrip(request); err != nil {
		return response, err
	} else {
		//todo 判断响应头context-type
		return response, err
	}
}
