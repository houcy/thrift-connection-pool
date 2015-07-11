#thrift_pool.go
golang thrift connection pool

## Usage
	var connectionPool *thrift_pool.ConnectionPool = thrift_pool.NewConnectionPool(CTR_THRIFT_POOL_SIZE, CTR_THRIFT_POOL_TIMEOUT, CTR_CLIENT_TIMES, example.CreateConnection, example.IsConnectionOpen, example.CloseConnection)
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