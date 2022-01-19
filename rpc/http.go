package rpc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/smallsung/gopkg/errors"
)

type httpError struct {
	errors.Err
	Status string
	Body   []byte
}

func newHttpError(status string, body []byte) *httpError {
	err := &httpError{Err: errors.NewErr("%v: %s", status, body), Status: status, Body: body}
	err.SetLocation(1)
	return err
}

func (err httpError) Error() string {
	if len(err.Body) == 0 {
		return err.Status
	}
	return fmt.Sprintf("%v: %s", err.Status, err.Body)
}

type httpServerConn struct {
	response http.ResponseWriter
	request  *http.Request
}

func (conn *httpServerConn) Read(p []byte) (n int, err error)  { return conn.request.Body.Read(p) }
func (conn *httpServerConn) Write(p []byte) (n int, err error) { return conn.response.Write(p) }
func (conn *httpServerConn) Close() error                      { return nil }

type httpClientConn uintptr

func (conn httpClientConn) Read(p []byte) (n int, err error)  { panic("implement me") }
func (conn httpClientConn) Write(p []byte) (n int, err error) { panic("implement me") }
func (conn httpClientConn) Close() error                      { panic("implement me") }

type httpClientCodec struct {
	ClientCodec
	url    *url.URL
	client *http.Client
}

func DialHTTPWithClient(ctx context.Context, url *url.URL, newCodecFunc NewClientCodecFunc, client *http.Client) *Client {
	c := &httpClientCodec{
		ClientCodec: newCodecFunc(new(httpClientConn)),
		url:         url,
		client:      client,
	}
	return NewClient(c)
}

func DialHTTP(ctx context.Context, url *url.URL, newCodecFunc NewClientCodecFunc) *Client {
	return DialHTTPWithClient(ctx, url, newCodecFunc, new(http.Client))
}
