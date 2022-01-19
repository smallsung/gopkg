package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/smallsung/gopkg/rpc"
	"github.com/smallsung/gopkg/rpc/jsonrpc"
	"go.uber.org/zap"
)

type service uintptr

func (receiver service) Subtract(a, b int) int    { return a - b }
func (receiver service) Update(a, b, c, d, e int) {}
func (receiver service) Foobar()                  {}

var logger, _ = zap.NewDevelopment()

const ipcEndpoint = "./ipcEndpoint"
const httpEndpoint = "127.0.0.1:11111"

var client *rpc.Client

func main() {
	var err error
	jsonRpcServer := rpc.NewServer(jsonrpc.NewServerCodec)
	jsonRpcServer.Logger = logger
	if err = jsonRpcServer.Register("rpc", new(service)); err != nil {
		panic(err)
	}

	//startUnix(jsonRpcServer)
	//time.Sleep(time.Second) //等待服务器启动
	//client = unixClient()
	//http
	startHttp(jsonRpcServer)
	time.Sleep(time.Second) //等待服务器启动
	client = httpClient()

	//--> data sent to Server
	//<-- data sent to Client
	call1()
	call2()
	call3()
	call4()
	call5()
	call6()
	call7()
	call8()
	call9()
	call10()
	call11()
	call12()

	select {}
}

func call12() {
	var result1 int
	var result2 int

	err := client.Bath(context.Background(), rpc.BatchElem{
		Method: "rpc.subtract",
		Params: []interface{}{42, 23},
		Result: &result1,
	}, rpc.BatchElem{
		Method: "rpc.subtract",
		Params: []interface{}{23, 42},
		Result: &result2,
	})
	if err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Printf("%+v\n%+v\n", result1, result2)
	}
}

//所有都为通知的rpc批量调用:
//--> [
//    {"jsonrpc": "2.0", "method": "notify_sum", "params": [1,2,4]},
//    {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]}
//]
//
//<-- //Nothing is returned for all notification batches
func call11() {
	//rpc.Client 不会出现部分情况。可以使用postman调试。
}

//rpc批量调用:
//--> [
//    {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
//    {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]},
//    {"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id": "2"},
//    {"foo": "boo"},
//    {"jsonrpc": "2.0", "method": "foo.get", "params": {"name": "myself"}, "id": "5"},
//    {"jsonrpc": "2.0", "method": "get_data", "id": "9"}
//    ]
//<-- [
//    {"jsonrpc": "2.0", "result": 7, "id": "1"},
//    {"jsonrpc": "2.0", "result": 19, "id": "2"},
//    {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//    {"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "5"},
//    {"jsonrpc": "2.0", "result": ["hello", 5], "id": "9"}
//    ]
func call10() {
	//rpc.Client 不会出现部分情况。可以使用postman调试。
}

//无效的rpc批量调用:
//--> [1,2,3]
//<-- [
//    {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//    {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//    {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//    ]
func call9() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//非空且无效的rpc批量调用:
//--> [1]
//<-- [
//    {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//    ]
func call8() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//包含空数组的rpc调用:
//--> []
//<-- {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
func call7() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//包含无效json的rpc批量调用:
//--> [
//        {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
//        {"jsonrpc": "2.0", "method"
//    ]
//<-- {"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}
func call6() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//包含无效请求对象的rpc调用:
//--> {"jsonrpc": "2.0", "method": 1, "params": "bar"}
//<-- {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
func call5() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//包含无效json的rpc调用:
//--> {"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]
//<-- {"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}
func call4() {
	//rpc.Client 不会出现这种情况。可以使用postman调试。
}

//不包含调用方法的rpc调用:
//--> {"jsonrpc": "2.0", "method": "foobar", "id": "1"}
//<-- {"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "1"}
func call3() {
	if err := client.Call(context.Background(), "", nil); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

//通知:
//--> {"rpc": "2.0", "method": "update", "params": [1,2,3,4,5]}
//--> {"rpc": "2.0", "method": "foobar"}
func call2() {
	if err := client.Notice(context.Background(), "rpc.update", 1, 2, 3, 4, 5); err != nil {
		fmt.Printf("%+v\n", err)
	}
	if err := client.Notice(context.Background(), "rpc.foobar"); err != nil {
		fmt.Printf("%+v\n", err)
	}
}

//带索引数组参数的rpc调用:
//--> {"rpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
//<-- {"rpc": "2.0", "result": 19, "id": 1}
//--> {"rpc": "2.0", "method": "subtract", "params": [23, 42], "id": 2}
//<-- {"rpc": "2.0", "result": -19, "id": 2}
func call1() {
	var result int
	if err := client.Call(context.Background(), "rpc.subtract", &result, 42, 23); err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Printf("%+v\n", result)
	}
	if err := client.Call(context.Background(), "rpc.subtract", &result, 23, 42); err != nil {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Printf("%+v\n", result)
	}
}

func startUnix(jsonRpcServer *rpc.Server) {
	listener, err := net.Listen("unix", ipcEndpoint)
	if err != nil {
		panic(err)
	}
	go jsonRpcServer.Accept(listener)
}

func startHttp(jsonRpcServer *rpc.Server) {
	go func() {
		if err := http.ListenAndServe(httpEndpoint, http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			jsonrpc.HttpHandlers.ValidHeader.ServeHTTP(response, request)
			jsonRpcServer.ServeHTTP(response, request)
		})); err != nil {
			panic(err)
		}
	}()
}

func httpClient() *rpc.Client {
	URL, err := url.Parse("http://" + httpEndpoint)
	if err != nil {
		panic(err)
	}
	client := rpc.DialHTTPWithClient(context.Background(), URL, jsonrpc.NewClientCodec, &http.Client{
		Transport:     jsonrpc.HttpRoundTripper.ValidHeader,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	})
	client.Logger = logger
	return client
}

func unixClient() *rpc.Client {
	client, err := rpc.DialIPC(context.Background(), ipcEndpoint, jsonrpc.NewClientCodec)
	if err != nil {
		panic(err)
	}
	client.Logger = logger
	return client
}
