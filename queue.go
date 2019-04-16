// ffmpeg

package ffmpeg

type QueueItem struct {
	InFile  string
	OutFile string
}

func NewQueueItem(inFile, outFile string) QueueItem {
	return QueueItem{InFile: inFile, OutFile: outFile}
}
