// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rpc

import (
	"context"
	"net"
)

func DialIPC(ctx context.Context, endpoint string, newCodecFunc NewClientCodecFunc) (*Client, error) {
	conn, err := new(net.Dialer).DialContext(ctx, "unix", endpoint)
	if err != nil {
		return nil, err
	}
	return NewClient(newCodecFunc(conn)), nil
}
