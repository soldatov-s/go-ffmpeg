// ffmpeg

package ffmpeg

import (
	"strings"

	"github.com/go-redis/redis"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	cl := &RedisClient{}
	cl.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	return cl
}

type RedisTask struct {
	QueueItem
	TaskName string
}

func (task *RedisTask) GetName() string {
	s := strings.Split(task.TaskName, ":")
	if len(s) > 1 {
		return s[1]
	}
	return ""
}

// Get from redis db task
// key as "tasks:taskID", where ID is INT type
// fields: inputfile, outfile, state.
// state: new - new task, busy - task in progress, completed - task completed
// holded - take to work
// redis-cli: hmset tasks:task1 outfile /home/ssoldatov/test/outfile.avi
//   inputfile /home/ssoldatov/test/inputfile.avi state new
func (cl *RedisClient) GetTask() (*RedisTask, error) {
	val, err := cl.client.Keys("tasks:*").Result()
	if err != nil {
		return nil, err
	}

	var result *RedisTask
	var key string
	txf := func(tx *redis.Tx) error {
		result = nil
		state, err := tx.HGet(key, "state").Result()
		if err != nil {
			return err
		}
		if state == "new" {
			inputfile, err := cl.client.HGet(key, "inputfile").Result()
			if err != nil {
				return err
			}
			outfile, err := cl.client.HGet(key, "outfile").Result()
			if err != nil {
				return err
			}
			result = &RedisTask{}
			result.InputFile = inputfile
			result.OutFile = outfile
			result.TaskName = key
			_, err = tx.HSet(key, "state", "holded").Result()
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, key = range val {
		err := cl.client.Watch(txf, key)
		if err == redis.TxFailedErr {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, nil
}

func (cl *RedisClient) BusyTask(task *RedisTask) error {
	_, err := cl.client.HSet(task.TaskName, "state", "busy").Result()
	if err != nil {
		return err
	}

	return nil
}

func (cl *RedisClient) CompleteTask(task *RedisTask) error {
	_, err := cl.client.HSet(task.TaskName, "state", "completed").Result()
	if err != nil {
		return err
	}

	return nil
}
