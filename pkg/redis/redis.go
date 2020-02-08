package redis

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

//Config def
type Config struct {
	Addr        string
	Password    string
	DB          int
	MaxIdel     int
	MaxActive   int
	IdleTimeout int
}

//NewRedisPool create a redids pool
func NewRedisPool(cfg *Config) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     cfg.MaxIdel,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.Addr, redis.DialPassword(cfg.Password), redis.DialDatabase(cfg.DB))
			if nil != err {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}


