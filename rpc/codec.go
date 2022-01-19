package rpc

import (
	"encoding/json"
	"io"
)

type jsonClientCodec struct {
	rwc     io.ReadWriteCloser
	encoder *json.Encoder
	decoder *json.Decoder
}

func (codec *jsonClientCodec) WriteRequest(raw json.RawMessage) error {
	return codec.encoder.Encode(raw)
}

func (codec *jsonClientCodec) ReadResponse() (json.RawMessage, error) {
	var raw json.RawMessage
	err := codec.decoder.Decode(&raw)
	return raw, err
}

func (codec *jsonClientCodec) Close() error {
	return codec.rwc.Close()
}

//type connCodec struct {
//	rwc io.ReadWriteCloser
//	*jsonServerCodec
//}
//
//func (codec *connCodec) Close() error {
//	return codec.rwc.Close()
//}
//
//func newConnServerCodec(conn io.ReadWriteCloser) ServerCodec {
//	return &connCodec{rwc: conn, jsonServerCodec: newJsonCodec(conn, conn)}
//}
//
//func newConnClientCodec(conn io.ReadWriteCloser) ClientCodec {
//	return &connCodec{rwc: conn, jsonServerCodec: newJsonCodec(conn, conn)}
//}
