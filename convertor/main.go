// main
package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/soldatov-s/go-ffmpeg"
)

func main() {
	fmt.Println("Start convertor")

	trc, err := ffmpeg.NewTranscoder()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	dbcl := ffmpeg.NewRedisClient("localhost:6379", "", 0)

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

	queue := make(chan *ffmpeg.RedisTask)
	next := make(chan *struct{})
	go func() {
		var i int
		for {
			task, err := dbcl.GetTask()
			if err != nil {
				fmt.Println("Error: ", err)
			}
			if task == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			fmt.Printf("Add new task%d\n", i)
			i++
			if i == math.MaxInt64 {
				i = 0
			}
			queue <- task
			next <- new(struct{})
		}
	}()

	go func() {
		for {
			fmt.Println("Wait new task...")
			v := <-queue
			fmt.Println("Getted new task")
			// Start transcoder process with progress checking
			fmt.Println("Input file:", v.InputFile)
			fmt.Println("Out file:", v.OutFile)
			done := trc.Run(v.InputFile, v.OutFile)

			progress := trc.Output()
			dbcl.BusyTask(v)

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
			dbcl.CompleteTask(v)
			fmt.Println()
			<-next
		}
	}()

	// Exit app if chan is closed
	<-exit

}
