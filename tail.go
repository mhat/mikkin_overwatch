package main

import (
	"github.com/ActiveState/tail"
	"log"
	"os"
	"time"
)

func attachTailer (msf *MikkinStreamFile, seek int64) {

	t, _ := tail.TailFile(msf.Path, tail.Config{
		Follow: true,
		ReOpen: true,
		Location: &tail.SeekInfo{-seek, os.SEEK_END} })
	msf.tail = t
	msf.Monitored = true

	log.Printf("Monitoring: %s", msf.Path)
	go func() {
		for {
			line := <-msf.tail.Lines
			msf.buffer.Value = line.Text
			msf.buffer = msf.buffer.Next()
			ChannelToWS <- WSServerMessage{msf.Path, line.Text}
		}
	}()

}


func monitorFiles () {
	duration, _ := time.ParseDuration("10s")

	for _, log := range OverwatchConfiguration.LogsToWatch.All() {
		msf := NewFile(log)
		MikkinStreamFiles = append(MikkinStreamFiles, *msf)
	}

	for {
		for i, _ := range MikkinStreamFiles {
			msf := &MikkinStreamFiles[i]

			// if the file is being monitored already we can ignore it
			if (msf.Monitored == true) {
				continue
			}

			// file's not moniterod, figure out a sensable amount to read at the tail end
			f, e := os.Open(msf.Path)
			if (os.IsNotExist(e)) {
				msf.Monitored = false
				continue
			}

			stat, _ := f.Stat()
			seek    := IntMin(stat.Size(), 1000)
			f.Close()

			// attach a tailer and mark msf as monitored
			attachTailer(msf, seek)
			msf.Monitored = true
		}

		// paaaaaause
		time.Sleep(duration)
	}

}
