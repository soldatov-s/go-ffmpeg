// main
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/soldatov-s/go-ffmpeg"
)

func main() {
	fmt.Println("Start convertor")

	trc, err := ffmpeg.NewTranscoder()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	stop := false
	exit := make(chan struct{})
	closeSignal := make(chan os.Signal)
	signal.Notify(closeSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeSignal
		trc.Stop()
		stop = true
		fmt.Println("\nExit program")
		close(exit)
	}()

	queue := make(chan ffmpeg.QueueItem)

	go func() {
		queue <- ffmpeg.NewQueueItem("/home/ssoldatov/test/inputfile.avi", "/home/ssoldatov/test/outfile.mp4")
	}()

	go func() {
		for {
			v := <-queue
			// Start transcoder process with progress checking
			fmt.Println("In file:", v.InFile, "Out file:", v.OutFile)
			done := trc.Run(v.InFile, v.OutFile)

			progress := trc.Output()

			for msg := range progress {
				if stop {
					return
				}
				fmt.Printf("\rConverted %3.2f%%", msg)
			}

			// Wait when transcoding process to end
			err = <-done
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
		}
	}()

	// Exit app if chan is closed
	<-exit

}
