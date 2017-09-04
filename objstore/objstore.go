// objstore stores key-value objects in a database. In this case it is a simple wrapper around redis
package objstore

import (
	"fmt"

	"time"

	"github.com/go-redis/redis"
)

// Store defines the methods needed for our concrete objstore service
type Store interface {
	Save(k string, v []byte, t int) error
	Close()
}

// Service is our redis implementation of the store interface
type Service struct {
	rd *redis.Client
}

// InitService creates a new redis backed service given the redis url
func InitService(rdURL string) (*Service, error) {
	opt, err := redis.ParseURL(rdURL)

	if err != nil {
		return &Service{}, fmt.Errorf("objstore/InitService: failed to parse options: %v", err)
	}

	c := redis.NewClient(opt)

	pong, err := c.Ping().Result()

	if err != nil || pong != "PONG" {
		return &Service{}, fmt.Errorf("objstore/InitService: failed to connect to redis: Ping-Pong result: %v Err: %v", pong, err)
	}

	return &Service{rd: c}, nil
}

// Save saves the given data to our objstore with the given timeout in seconds, 0 seconds for no timeout
func (s *Service) Save(k string, v []byte, t int) error {
	return s.rd.Set(k, v, time.Duration(t)*time.Second).Err()
}

// Close closes our underlying redis connection
func (s *Service) Close() {
	s.rd.Close()
}
