// ffmpeg

package ffmpeg

import (
	"encoding/json"
	//	"fmt"

	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	cl := &RedisClient{}
	cl.client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return cl
}

type RedisTask struct {
	QueueItem
	state string
}

type RedisTasks struct {
	tasks []RedisTask
}

// MarshalBinary -
func (e *RedisTasks) MarshalBinary() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalBinary -
func (e *RedisTasks) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}

	return nil
}

func (cl *RedisClient) GetTask() (*QueueItem, error) {
	val, err := cl.client.Keys("tasks:*").Result()
	if err != nil {
		return nil, err
	}

	for _, k := range val {
		state, err := cl.client.HGet(k, "state").Result()
		if err != nil {
			return nil, err
		}
		if state == "new" {
			inputfile, err := cl.client.HGet(k, "inputfile").Result()
			if err != nil {
				return nil, err
			}
			outfile, err := cl.client.HGet(k, "outfile").Result()
			if err != nil {
				return nil, err
			}
			return NewQueueItem(inputfile, outfile), nil
		}
	}
	return nil, nil
}
