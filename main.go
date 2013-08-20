package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/ActiveState/tail"
	"log"
	"fmt"
	"net/http"
	"container/ring"
	"flag"
)


type MikkinStreamFile struct {
	Path string
	Name string
	Description string
	Monitored bool
	tail *tail.Tail
	buffer *ring.Ring
}
var MikkinStreamFiles []MikkinStreamFile

func NewFile (log LogToWatch) *MikkinStreamFile {
	return &MikkinStreamFile{
		Name: log.Name,
		Path: log.Path,
		Description: log.Description,
		Monitored: false,
		tail: nil,
		buffer: ring.New(10)}
}

func channelToWSReader () {
	for {
		// wait for a logged message
		msg := <-ChannelToWS

		// no clients, no problem, move-on!
		if (len(ClientConnections) == 0) {
			continue
		}

		// send it to all the clients
		for _, client := range ClientConnections {
			_, filter := client.filters[msg.Channel]
			if filter == true {
				log.Printf("Rejecting message on channel %s for client %s\n", msg.Channel, client.clientIP)
				continue
			}
			if err := websocket.JSON.Send(client.websocket, msg); err != nil {
				// we could not send the message to a peer
				log.Print("Error sending to client %s. %s\n", client.clientIP, err.Error())
			}
		}
	}
}


func init() {
	flag.StringVar(&OverwatchConfiguration.ServerConfigFile, "config", "config/server.json", "/path/to/config")
}

func main() {
	flag.Parse()
	OverwatchConfiguration.LoadConfiguration()

	go monitorFiles()
	go channelToWSReader()

	log.Printf("------------------------------\n")
	log.Printf("Register Static Assets\n")
	http.Handle("/assets/",   http.FileServer(http.Dir(OverwatchConfiguration.RootPath)))

	log.Printf("Registering Views\n")
	http.HandleFunc("/",      HomeHandler)

	log.Printf("Registering Resource\n")
	http.HandleFunc("/logs",  LogsHandler)

	log.Printf("Register WebSocket Handler\n")
	http.Handle("/websocket", websocket.Handler(SockServer))
	log.Printf("------------------------------\n")

	err := http.ListenAndServe(fmt.Sprintf("%s:%d",
		OverwatchConfiguration.BindAddress,
		OverwatchConfiguration.BindPort), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
