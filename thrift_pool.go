package thrift_pool

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type Connection struct {
	client     interface{}
	createTime time.Time //创建时间
}

type ConnectionPool struct {
	pool_size int           //池容量
	timeout   time.Duration //client过期时间

	mu sync.Mutex

	activeList   *list.List //正在使用的连接
	inactiveList *list.List //空闲的连接

	createConnection func() (interface{}, error)
	isConnectionOpen func(client interface{}) bool
	closeConnection  func(client interface{}) error
}

//创建一个连接池
func NewConnectionPool(
	pool_size int,
	timeout time.Duration,
	createConnection func() (interface{}, error),
	isConnectionOpen func(client interface{}) bool,
	closeConnection func(client interface{}) error) *ConnectionPool {

	p := &ConnectionPool{
		pool_size:        pool_size,
		timeout:          timeout,
		activeList:       list.New(),
		inactiveList:     list.New(),
		createConnection: createConnection,
		isConnectionOpen: isConnectionOpen,
		closeConnection:  closeConnection,
	}
	return p
}

//得到一个client链接
func (p *ConnectionPool) GetConnection(clientChan chan interface{}, errChan chan error) {
	//满了
	if p.activeList.Len() >= p.pool_size {
		errChan <- errors.New("connection pool is full")
		//检查是否有过期的
		go func() {
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
		}()
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
	connection := &Connection{client: client, createTime: time.Now()}
	p.mu.Lock()
	p.activeList.PushBack(connection)
	p.mu.Unlock()
	clientChan <- connection.client
	return

}

//将用完的client放回list
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
	//有可能太晚归还 已经被删除
	if findInActiveList {
		connection.createTime = time.Now()
		p.mu.Lock()
		p.inactiveList.PushBack(connection)
		p.mu.Unlock()
	}

}

//错误的client，直接删去
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
