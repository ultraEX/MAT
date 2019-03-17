// redisgo
package use_redis

import (
	"fmt"
	"time"

	"../../config"

	"sync"

	"github.com/gomodule/redigo/redis"
)

const (
	DEFAULT_POOL_MAX_IDLE      = 1
	DEFAULT_POOL_MAX_ACTIVE    = 50
	DEFAULT_POOL_IDLE_TIMEOUT  = 180 * time.Second
	DEFAULT_POOL_MAX_LIFE_TIME = 60 * time.Second
)

type LongConn struct {
	long     *redis.Conn
	ConMutex sync.Mutex
}

type RediGODB struct {
	pool *redis.Pool
	LongConn
}

func NewRediGODB() *RediGODB {

	pool := redis.Pool{
		MaxIdle:         DEFAULT_POOL_MAX_IDLE,
		MaxActive:       DEFAULT_POOL_MAX_ACTIVE,
		IdleTimeout:     DEFAULT_POOL_IDLE_TIMEOUT,
		MaxConnLifetime: DEFAULT_POOL_MAX_LIFE_TIME,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", config.GetMEConfig().Redis_MX_IP, config.GetMEConfig().Redis_MX_Port))
			if err != nil {
				return nil, err
			}
			if len(config.GetMEConfig().Redis_MX_Pwd) > 0 {
				if _, err := c.Do("AUTH", config.GetMEConfig().Redis_MX_Pwd); err != nil {
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", config.GetMEConfig().Redis_MX_Db); err != nil {
				return nil, err
			}
			return c, nil
		},

		// 用来测试连接是否可用
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	o := new(RediGODB)
	o.pool = &pool

	conn := o.GetConn()
	o.LongConn.long = &conn

	return o
}

// 关闭redis管理对象，将会关闭底层的
func (t *RediGODB) Close() error {
	return t.pool.Close()
}

// 获得一个原生的redis连接对象，用于自定义连接操作，
// 但是需要注意的是如果不再使用该连接对象时，需要手动Close连接，否则会造成连接数超限。
func (t *RediGODB) GetConn() redis.Conn {
	return t.pool.Get()
}

// 执行同步命令 - Do
func (t *RediGODB) Do(command string, args ...interface{}) (interface{}, error) {
	conn := t.pool.Get()
	defer conn.Close()
	return conn.Do(command, args...)
}

// 执行异步命令 - Send
func (t *RediGODB) Send(command string, args ...interface{}) error {
	conn := t.pool.Get()
	defer conn.Close()
	return conn.Send(command, args...)
}

// 执行同步命令 - Do
func (t *RediGODB) LongConnDo(command string, args ...interface{}) (interface{}, error) {
	conn := *t.LongConn.long
	t.LongConn.ConMutex.Lock()
	defer t.LongConn.ConMutex.Unlock()
	return conn.Do(command, args...)
}

// 执行异步命令 - Send
func (t *RediGODB) LongConnSend(command string, args ...interface{}) error {
	conn := *t.LongConn.long
	t.LongConn.ConMutex.Lock()
	defer t.LongConn.ConMutex.Unlock()
	return conn.Send(command, args...)
}
