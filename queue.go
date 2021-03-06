// ffmpeg

package ffmpeg

type QueueItem struct {
	InputFile string
	OutFile   string
}

func NewQueueItem(inFile, outFile string) *QueueItem {
	return &QueueItem{InputFile: inFile, OutFile: outFile}
}

func (item *QueueItem) Copy() *QueueItem {
	return &QueueItem{InputFile: item.InputFile, OutFile: item.OutFile}
}
