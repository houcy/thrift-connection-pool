package example

import (
	_ "fmt"
	"gen-go/hello"
	"git.apache.org/thrift.git/lib/go/thrift"
	"net"
)

const (
	CTR_HOST      = "127.0.0.1"
	CTR_HOST_PORT = "19090"
)

//创建一个thrift的client
func CreateConnection() (interface{}, error) {
	var client *hello.HelloClient
	var transport thrift.TTransport
	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	raw_transport, e1 := thrift.NewTSocket(net.JoinHostPort(CTR_HOST, CTR_HOST_PORT))
	if e1 != nil {
		return client, e1
	}
	transport = transportFactory.GetTransport(raw_transport)
	e2 := transport.Open()
	if e2 != nil {
		return client, e2
	}

	client = hello.NewHelloClientFactory(transport, protocolFactory)
	return client, nil
}

//连接是否正常
func IsConnectionOpen(client interface{}) bool {
	return client.(*hello.HelloClient).Transport.IsOpen()
}

//关闭
func CloseConnection(client interface{}) error {
	return client.(*hello.HelloClient).Transport.Close()
}
