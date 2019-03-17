package use_redis

import (
	"github.com/gomodule/redigo/redis"
)

//---------------------------------------------------------------------------------------------------
///	ZADD key score1 member1 [score2 member2]
func (t *RedisDb) ZSetAddInt64(key string, score float64, mem int64) {
	_, err := t.Do("ZADD", key, score, mem)
	if err != nil {
		panic(err)
	}
}

/// ZREM key member [member ...]
func (t *RedisDb) ZSetRemoveInt64(key string, mem int64) {
	_, err := t.Do("ZREM", key, mem)
	if err != nil {
		panic(err)
	}
}

/// ZRANGE key start stop [WITHSCORES]
func (t *RedisDb) ZSetGetRangeInt64s(key string, start int64, stop int64) (v []int64) {
	v, err := redis.Int64s(t.Do("ZRANGE", key, start, stop))
	if err != nil {
		panic(err)
	}
	return v
}

/// ZRANGE key start stop [WITHSCORES]
func (t *RedisDb) ZSetGetInt64(key string, index int64) (int64, bool) {
	// startTime := time.Now().UnixNano()

	v, err := redis.Int64s(t.Do("ZRANGE", key, index, index))
	if err != nil {
		panic(err)
	}

	// endTime := time.Now().UnixNano()
	// fmt.Printf("DO waist time: %f second\n", float64(endTime-startTime)/float64(1*time.Second))

	if len(v) == 1 {
		return v[0], true
	} else {
		return -1, false
	}
}

/// ZRANGE key start stop [WITHSCORES]
/// ZCARD key
func (t *RedisDb) ZSetGetAll(key string) []int64 {
	count, err := redis.Int64(t.Do("ZCARD", key))
	if err != nil {
		panic(err)
	}

	if count > 0 {
		v, err := redis.Int64s(t.Do("ZRANGE", key, 0, count-1))
		if err != nil {
			panic(err)
		}
		return v
	} else {
		return nil
	}
}

func (t *RedisDb) ZSetGetSize(key string) int64 {
	count, err := redis.Int64(t.Do("ZCARD", key))
	if err != nil {
		panic(err)
	}
	return count
}
