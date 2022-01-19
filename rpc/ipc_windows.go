//+build windows

package rpc

import "context"

func DialIPC(ctx context.Context, endpoint string, newCodecFunc NewClientCodecFunc) (*Client, error) {
	panic("")
}
