// ffmpeg

package ffmpeg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type Transcoder struct {
	stdErrPipe   io.ReadCloser
	stdStdinPipe io.WriteCloser
	process      *exec.Cmd
	binpath      string
}

func NewTranscoder() (*Transcoder, error) {
	var out bytes.Buffer
	trc := &Transcoder{}

	cmd := exec.Command("which", "ffmpeg")
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	trc.binpath = strings.Replace(out.String(), "\n", "", -1)

	return trc, err
}

func (trc *Transcoder) Run(inPath, outPath string) <-chan error {
	done := make(chan error)
	args := []string{"-i", inPath, "-c:v", "libx264", "-b:v",
		"1000k", "-c:a", "aac", "-f", "mp4", outPath, "-y"}

	proc := exec.Command(trc.binpath, args...)
	errStream, err := proc.StderrPipe()
	if err != nil {
		fmt.Println("Process not available: " + err.Error())
	} else {
		trc.stdErrPipe = errStream
	}

	stdin, err := proc.StdinPipe()
	if nil != err {
		fmt.Println("Stdin not available: " + err.Error())
	}

	trc.stdStdinPipe = stdin

	out := &bytes.Buffer{}
	proc.Stdout = out

	err = proc.Start()

	trc.process = proc
	go func(err error, out *bytes.Buffer) {
		if err != nil {
			done <- fmt.Errorf("Failed Start FFMPEG (%s) with %s, message %s", args, err, out.String())
			close(done)
			return
		}
		err = proc.Wait()
		if err != nil {
			err = fmt.Errorf("Failed Finish FFMPEG (%s) with %s message %s", args, err, out.String())
		}
		done <- err
		close(done)
	}(err, out)

	return done
}

func GetFieldValue(line, fieldsep, fieldname, valuesep string) string {
	params := strings.Split(line, fieldsep)
	for _, p := range params {
		if strings.Contains(p, fieldname) {
			fieldSplit := strings.Split(strings.Trim(p, " "), valuesep)
			if len(fieldSplit) > 1 {
				return fieldSplit[1]
			}
		}
	}
	return ""
}

func (trc Transcoder) Output() <-chan float64 {
	out := make(chan float64)

	go func() {
		defer close(out)
		if trc.stdErrPipe == nil {
			out <- -1
			return
		}

		defer trc.stdErrPipe.Close()

		scanner := bufio.NewScanner(trc.stdErrPipe)

		split := func(data []byte, atEOF bool) (advance int, token []byte, spliterror error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				// We have a full newline-terminated line.
				return i + 1, data[0:i], nil
			}
			if i := bytes.IndexByte(data, '\r'); i >= 0 {
				// We have a cr terminated line
				return i + 1, data[0:i], nil
			}
			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		}

		scanner.Split(split)
		buf := make([]byte, 2)
		scanner.Buffer(buf, bufio.MaxScanTokenSize)
		dursec := float64(-1)

		for scanner.Scan() {
			line := scanner.Text()
			if dursec == -1 {
				if strings.Contains(line, "Duration:") && strings.Contains(line, "start:") && strings.Contains(line, "bitrate:") {
					Duration := GetFieldValue(line, ",", "Duration", " ")
					dursec = DurToSec(Duration)
				}
			}

			if strings.Contains(line, "frame=") && strings.Contains(line, "time=") && strings.Contains(line, "bitrate=") {
				currentTime := GetFieldValue(line, " ", "time", "=")

				timesec := DurToSec(currentTime)
				//live stream check
				if dursec != 0 {
					// Progress calculation
					out <- (timesec * 100) / dursec
				} else {
					out <- -1
				}
			}
		}
	}()

	return out
}

func (t *Transcoder) Stop() error {
	if t.process != nil {
		stdin := t.stdStdinPipe
		if stdin != nil {
			stdin.Write([]byte("q\n"))
		}
	}
	return nil
}

func (trc *Transcoder) PrintInfo() {
	fmt.Println(trc.binpath)
}

func DurToSec(dur string) (sec float64) {
	durAry := strings.Split(dur, ":")
	var secs float64
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.ParseFloat(durAry[0], 64)
	secs = hr * (60 * 60)
	min, _ := strconv.ParseFloat(durAry[1], 64)
	secs += min * (60)
	second, _ := strconv.ParseFloat(durAry[2], 64)
	secs += second
	return secs
}
