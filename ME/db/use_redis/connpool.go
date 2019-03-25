package use_redis

import (
	"github.com/gomodule/redigo/redis"
)

type ConnPool struct {
	Size      int
	redispool *RediGODB
	idleQueue chan *redis.Conn
}

func NewConnPool(size int) *ConnPool {
	o := new(ConnPool)
	o.Size = size
	o.redispool = NewRediGODB()
	o.idleQueue = make(chan *redis.Conn, o.Size)
	for i := 0; i < o.Size; i++ {
		conn := o.redispool.GetConn()
		o.idleQueue <- &conn
	}
	return o
}

func (t *ConnPool) GetConn() *redis.Conn {
	return <-t.idleQueue
}

func (t *ConnPool) GetNamedLongConn(name string) *LongConn {
	return t.redispool.GetLongConn(name)
}

func (t *ConnPool) RecycleConn(conn *redis.Conn) {
	t.idleQueue <- conn
}

func (t *ConnPool) Do(command string, args ...interface{}) (interface{}, error) {
	conn := t.GetConn()
	defer t.RecycleConn(conn)
	return (*conn).Do(command, args...)
}

func (t *ConnPool) Send(command string, args ...interface{}) error {
	conn := t.GetConn()
	defer t.RecycleConn(conn)
	return (*conn).Send(command, args...)
}
