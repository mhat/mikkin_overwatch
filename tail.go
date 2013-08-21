package main

import (
	"github.com/ActiveState/tail"
	"log"
	"os"
	"time"
	"container/ring"
)

type WatchedLogFile struct {
	Info LogToWatch
	TailerIsAttached bool
	Buffer *ring.Ring
}
var WatchedLogFiles []WatchedLogFile

func NewWatchedLogFile (log LogToWatch) *WatchedLogFile {
	logfile := WatchedLogFile{log, false, ring.New(10)}
	WatchedLogFiles = append(WatchedLogFiles, logfile)
	return &logfile
}

func FileWatcher (logsToWatch []LogToWatch, broadcast chan WSServerMessage) {
	duration, _ := time.ParseDuration("10s")
	for _, log := range logsToWatch {
		NewWatchedLogFile(log)
	}

	for {
		for i, _ := range WatchedLogFiles {
			wlog := &WatchedLogFiles[i]

			if (wlog.TailerIsAttached == true) {
				continue
			}

			// check to see if the file exists
			wlogfh, e := os.Open(wlog.Info.Path)
			if (os.IsNotExist(e)) {
				wlog.TailerIsAttached = false
				continue
			}

			// now we want to pre-populate the log-buffer with the last few bytes
			// of the log file; this makes it a little nicer for clients on their
			// initial connection
			stat, _ := wlogfh.Stat()
			seek    := IntMin(stat.Size(), 2000)
			wlogfh.Close()

			log.Printf("Attching Tailer to %s\n", wlog.Info.Path)
			tailer, _ := tail.TailFile(wlog.Info.Path, tail.Config{
				Follow: true,
				ReOpen: true,
				Location: &tail.SeekInfo{-seek, os.SEEK_END} })

			wlog.TailerIsAttached = true
			go func() {
				for {
					newline := <-tailer.Lines
					//log.Printf("Read from [%s] text [%s]\n", wlog.Info.Path, newline.Text)
					wlog.Buffer.Value = newline.Text
					wlog.Buffer = wlog.Buffer.Next()
					broadcast <- WSServerMessage{wlog.Info.Path, newline.Text}
				}
			}()

		}

		// paaaaaause
		time.Sleep(duration)
	}

}

