package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/ActiveState/tail"
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"encoding/json"
)

const listenAddr = "localhost:4000"
var   pwd, _     = os.Getwd()


type MikkinStreamFile struct {
	Path string
	Name string
	Description string
	Monitored bool 
	tail *tail.Tail
}
var MikkinStreamFiles []MikkinStreamFile


type WSClientMessage struct {
	Channel string
	Message string
}

type WSServerMessage struct {
	Channel string
	Content string
}

var ChannelToWS = make(chan WSServerMessage)


type ClientConnection struct {
	websocket *websocket.Conn
	clientIP  string
}
var ClientConnections = make(map[ClientConnection]int)


func channelToWSReader () {
	for {
		msg := <-ChannelToWS

		// no clients, no problem, move-on!
		if (len(ClientConnections) == 0) {
			continue
		}

		// send it to all the clients
		for client := range ClientConnections {
			log.Printf("sent %d bytes to client %s on channel %s", len(msg.Content), client.clientIP, msg.Channel)
			if err := websocket.JSON.Send(client.websocket, msg); err != nil {
				// we could not send the message to a peer
				log.Println("could not send message to ", client.clientIP, err.Error())
			}
		}
	}
}


type ConfigFileEntry struct {
	Path string
        Name string
        Description string
}

func readConfig() {
	var entries []ConfigFileEntry
	sysconf, _ := ioutil.ReadFile("./config/system.json")
	e := json.Unmarshal(sysconf, &entries)

	if (e != nil) {
		log.Printf("Sorry! config/system.json appears to be invalid!\n>> %s\n\n", e)
		os.Exit(1)
	}

	for _, cfe := range entries {
		log.Printf("Adding %s per config/system.json configuration.\n", cfe.Path)
		MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{
			Name: cfe.Name,
			Description: cfe.Description,
			Path: cfe.Path,
			Monitored: false,
			tail: nil})
	}

	var svclist []string;
	svcconf, _ := ioutil.ReadFile("./config/service.json")
	json.Unmarshal(svcconf, &svclist)

	if (e != nil) {
		log.Printf("Sorry! config/service.json appears to be invalid!\n>> %s\n\n", e)
		os.Exit(1)
	}

	for _, svc := range svclist {
		log.Printf("Adding service logs for %s\n", svc)
		MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{
			fmt.Sprintf("/var/log/%s/%s.log", svc),
			fmt.Sprintf("%s-service-log", svc),
			fmt.Sprintf("%s service log", svc), false, nil})

		MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{
			fmt.Sprintf("/var/log/%s/gc.log", svc),
			fmt.Sprintf("%s-gc-log", svc),
			fmt.Sprintf("%s garbage collector log", svc), false, nil})

		MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{
			fmt.Sprintf("/var/log/%s/requests.log", svc),
			fmt.Sprintf("%s-request-log", svc),
			fmt.Sprintf("%s request log", svc), false, nil})

	}
}

func main() {
	log.Printf("Initialize\n")
	readConfig();
	MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{ "/var/log/system.log", "system", "systems generic logfile", false, nil })
	MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{ "/var/log/wifi.log",   "wifi",   "systems wifi log"       , false, nil })
	MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{ "/tmp/knopp.log",      "knopp",  "knopp log"              , false, nil })
	MikkinStreamFiles = append(MikkinStreamFiles, MikkinStreamFile{ "/tmp/nooop.log",      "nope",   "not a log yet"          , false, nil })

	go monitorFiles()
	go channelToWSReader()

	log.Printf("Register Static Assets\n")
	http.Handle("/assets/", http.FileServer(http.Dir(pwd)))

	log.Printf("Registering Views\n")
	http.HandleFunc("/",        HomeHandler)

	log.Printf("Registering Resource\n")
	http.HandleFunc("/logs",    LogsHandler)

	log.Printf("Register WebSocket Handler\n")
	http.Handle("/websocket", websocket.Handler(SockServer))


	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
