package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

type RedisWrapper struct {
	c redis.Conn
}

//localhost:6379
func (r *RedisWrapper) Connect(server string) (err error) {
	r.c, err = redis.Dial("tcp", server)
	if err != nil {
		log.Printf("conn redis failed , err: %v", err)
		return err
	}
	return nil
}

func (r *RedisWrapper) SetKey(key string, value string) error {

	_, err := r.c.Do("Set", key, value)
	if err != nil {
		log.Printf("Set [%s] [%s]  error  %v", key, value, err)
		return err
	}
	return nil

}

func (r *RedisWrapper) AppendKey(key string, value string) error {

	_, err := r.c.Do("append", key, value)
	if err != nil {
		log.Printf("Append [%s] [%s]  error  %v", key, value, err)
		return err
	}
	return nil

}

func (r *RedisWrapper) GetKey(key string) (string, error) {

	res, err := redis.String(r.c.Do("Get", key))
	if err != nil {
		log.Printf("Get [%s]  error  %v", key, err)
		return "", err
	}
	return res, nil

}
