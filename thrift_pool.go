package thrift_pool

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type Connection struct {
	client      interface{}
	createTime  time.Time
	calledTimes int64
}

type ConnectionPool struct {
	pool_size int
	timeout   time.Duration

	mu sync.Mutex

	activeList   *list.List
	inactiveList *list.List

	clientTimes int64 //the connection will be closed after clientTimes,0 is unlimited

	createConnection func() (interface{}, error)
	isConnectionOpen func(client interface{}) bool
	closeConnection  func(client interface{}) error
}

//创建一个连接池
func NewConnectionPool(
	pool_size int,
	timeout time.Duration,
	clientTimes int64,
	createConnection func() (interface{}, error),
	isConnectionOpen func(client interface{}) bool,
	closeConnection func(client interface{}) error) *ConnectionPool {

	p := &ConnectionPool{
		pool_size:        pool_size,
		timeout:          timeout,
		activeList:       list.New(),
		inactiveList:     list.New(),
		clientTimes:      clientTimes,
		createConnection: createConnection,
		isConnectionOpen: isConnectionOpen,
		closeConnection:  closeConnection,
	}
	return p
}

//get a connection from pool
func (p *ConnectionPool) GetConnection(clientChan chan interface{}, errChan chan error) {
	//满了
	if p.activeList.Len() >= p.pool_size {
		errChan <- errors.New("connection pool is full")
		//check expired
		var next *list.Element
		now := time.Now()
		var diff time.Duration
		for e := p.activeList.Front(); e != nil; e = next {
			next = e.Next()
			connection := e.Value.(*Connection)
			diff = now.Sub(connection.createTime)
			if diff >= p.timeout {
				p.closeConnection(connection.client)
				p.mu.Lock()
				p.activeList.Remove(e)
				p.mu.Unlock()
			}
		}
		return
	}

	var next *list.Element
	for e := p.inactiveList.Front(); e != nil; e = next {
		next = e.Next()
		connection := e.Value.(*Connection)
		client := connection.client
		if p.isConnectionOpen(client) {
			clientChan <- client
			p.mu.Lock()
			p.inactiveList.Remove(e)
			p.activeList.PushBack(connection)
			p.mu.Unlock()
			return
		} else {
			p.mu.Lock()
			p.inactiveList.Remove(e)
			p.mu.Unlock()
		}
	}

	client, e1 := p.createConnection()
	if e1 != nil {
		errChan <- e1
		return
	}
	connection := &Connection{client: client, createTime: time.Now(), calledTimes: 0}
	p.mu.Lock()
	p.activeList.PushBack(connection)
	p.mu.Unlock()
	clientChan <- connection.client
	return

}

//return the connection to the pool
func (p *ConnectionPool) ReturnConnection(client interface{}) {
	var connection *Connection
	var next *list.Element
	var findInActiveList bool = false
	for e := p.activeList.Front(); e != nil; e = next {
		next = e.Next()
		connection = e.Value.(*Connection)
		if client == connection.client {
			p.mu.Lock()
			findInActiveList = true
			p.activeList.Remove(e)
			p.mu.Unlock()
			break
		}
	}
	if findInActiveList {
		if p.clientTimes > 0 && connection.calledTimes >= p.clientTimes {
			p.closeConnection(client)
			return
		}
		connection.calledTimes++
		connection.createTime = time.Now()
		p.mu.Lock()
		p.inactiveList.PushBack(connection)
		p.mu.Unlock()
	}

}

func (p *ConnectionPool) ReportErrorConnection(client interface{}) {
	var next *list.Element
	for e := p.activeList.Front(); e != nil; e = next {
		next = e.Next()
		connection := e.Value.(*Connection)
		if client == connection.client {
			p.closeConnection(client)
			p.mu.Lock()
			p.activeList.Remove(e)
			p.mu.Unlock()
			break
		}
	}
}
