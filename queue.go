// ffmpeg

package ffmpeg

import (
	"sync"
)

type QueueItem struct {
	InFile  string
	OutFile string
}

func NewQueueItem(inFile, outFile string) *QueueItem {
	return &QueueItem{InFile: inFile, OutFile: outFile}
}

type Queue struct {
	Items map[int]*QueueItem
	mu *sync.Mutex
}

func NewQueue() *Queue {
	obj := &Queue{}
	obj.Items = make(map[int]*QueueItem)
	obj.mu = &sync.Mutex{}
	return obj
}

func (obj *Queue) Push(key int, item *QueueItem) {
	defer obj.mu.Unlock()
	obj.mu.Lock()
	obj.Items[key] = item
}

func (obj *Queue) Remove(key int) {
	defer obj.mu.Unlock()
	obj.mu.Lock()
	item := obj.Items[key]
	if item == nil {
		return
	}
	delete(obj.Items, key)

}
