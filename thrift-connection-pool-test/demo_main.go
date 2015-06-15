package main

import (
	"example"
	"fmt"
	"gen-go/hello"
	"github.com/yilee/thrift-connection-pool"
	"time"
)

const (
	CTR_TIME_WAIT           = time.Second * 1  //thrift服务超时
	CTR_THRIFT_POOL_SIZE    = 3                //池容量
	CTR_THRIFT_POOL_TIMEOUT = time.Minute * 15 //15分钟，连接池里的client过期时间
)

func main() {
	var connectionPool *thrift_pool.ConnectionPool = thrift_pool.NewConnectionPool(CTR_THRIFT_POOL_SIZE, CTR_THRIFT_POOL_TIMEOUT, example.CreateConnection, example.IsConnectionOpen, example.CloseConnection)

	var client *hello.HelloClient
	clientChan := make(chan interface{})
	errClientChan := make(chan error)

	go connectionPool.GetConnection(clientChan, errClientChan)
	select {
	case res := <-clientChan:
		client = res.(*hello.HelloClient)
	case err := <-errClientChan:
		fmt.Println("error", err)
		return
	case <-time.After(CTR_TIME_WAIT):
		fmt.Println("获取client超时")
		return
	}

	res, err := client.HelloString("Hello")
	if err != nil {
		fmt.Println("err", err)
		connectionPool.ReportErrorConnection(client)
	} else {
		fmt.Println(res)
		connectionPool.ReturnConnection(client)
	}

}
