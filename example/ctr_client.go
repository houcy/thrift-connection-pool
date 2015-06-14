package example

import (
	"git.apache.org/thrift.git/lib/go/thrift"
	"net"
)

const (
	CTR_HOST      = "pickad.glorfindel.daesvc.douban.com"
	CTR_HOST_PORT = "7303"
)

//创建一个thrift的client
func CreateConnection() (interface{}, error) {
	var client *PickAdClient
	var transport thrift.TTransport
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
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
	message := GetCTRHelloMessage()
	_, e3 := raw_transport.Write(message)
	if e3 != nil {
		return client, e3
	}
	client = NewPickAdClientFactory(transport, protocolFactory)
	return client, nil
}

//连接是否正常
func IsConnectionOpen(client interface{}) bool {
	return client.(*PickAdClient).Transport.IsOpen()
}

//关闭
func CloseConnection(client interface{}) error {
	return client.(*PickAdClient).Transport.Close()
}
