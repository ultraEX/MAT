package use_redis

import (
	"sync"
)

const (
	CONNPOOL_SIZE = 20
)

type RedisDb struct {
	*ConnPool
}

var redisDbObj *RedisDb
var once sync.Once

func RedisDbInstance() *RedisDb {

	once.Do(func() {
		redisDbObj = new(RedisDb)
		redisDbObj.ConnPool = NewConnPool(CONNPOOL_SIZE)
	})

	return redisDbObj
}
