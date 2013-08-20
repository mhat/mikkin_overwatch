package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/ActiveState/tail"
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"net/url"
	"encoding/json"
	"container/ring"
	"flag"
)

const listenAddr = "0.0.0.0:4000"
var   pwd, _     = os.Getwd()


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


type LogToWatch struct {
	Path string
        Name string
        Description string
}

type LogsToWatchStruct struct {
	StandardLogs []LogToWatch
	DropWizardServices []string
	DropWizardDotDPath string
	DropWizardDotDServices []string
}

func (logs *LogsToWatchStruct) All() []LogToWatch {
	watchList := make([]LogToWatch, len(logs.StandardLogs))
	copy(watchList, logs.StandardLogs)

	ftso := map[string]bool{}

	for _, svc := range logs.DropWizardServices {
		ftso[svc] = true
	}
	for _, svc := range logs.DropWizardDotDServices {
		ftso[svc] = true
	}

	for svc, _ := range ftso {
		watchList = append(watchList, LogToWatch{
			Path: fmt.Sprintf("/var/log/%s/%s.log", svc, svc),
			Name: fmt.Sprintf("%s-service-log", svc),
			Description: fmt.Sprintf("%s service log", svc)})

		watchList = append(watchList, LogToWatch{
			Path: fmt.Sprintf("/var/log/%s/gc.log", svc),
			Name: fmt.Sprintf("%s-gc-log", svc),
			Description: fmt.Sprintf("%s garbage collector log", svc)})

		watchList = append(watchList, LogToWatch{
			Path: fmt.Sprintf("/var/log/%s/requests.log", svc),
			Name: fmt.Sprintf("%s-request-log", svc),
			Description: fmt.Sprintf("%s request log", svc)})
	}

	return watchList
}

type OverwatchConfigurationStruct struct {
	ServerConfigFile string
	LogsToWatchConfigPath string

	ServerName string
	BindAddress string
	BindPort int64
	WebSocketUrl *url.URL

	TemplatePath string
	AssetPath string

	Verbose bool
	RootPath string

	LogsToWatch LogsToWatchStruct
}
var OverwatchConfiguration OverwatchConfigurationStruct

func loadConfiguration() {
	file, err1 := os.Open(OverwatchConfiguration.ServerConfigFile)
	if (err1 != nil) {
		log.Printf("Error: %s\n", err1)
		os.Exit(-1)
	}

	file.Close()
	jsonBlob, _ := ioutil.ReadFile(OverwatchConfiguration.ServerConfigFile)

	err2 := json.Unmarshal(jsonBlob, &OverwatchConfiguration)
	if (err2 != nil) {
		log.Printf("Error Parsing Configuration: %s\n", err2)
		os.Exit(-1)
	}

	OverwatchConfiguration.RootPath, _ = os.Getwd()
	OverwatchConfiguration.WebSocketUrl, _ = url.Parse(fmt.Sprintf("ws://%s:%d/websocket",
		OverwatchConfiguration.ServerName,
		OverwatchConfiguration.BindPort))

	dwDir, err3 := ioutil.ReadDir(OverwatchConfiguration.LogsToWatch.DropWizardDotDPath)
	if (err3 != nil) {
		log.Printf("Error: %s\n", err3)
	}

	var dwDirServices []string
	for _, dwInfo := range dwDir {
		dwDirServices = append(dwDirServices, dwInfo.Name())
	}
	OverwatchConfiguration.LogsToWatch.DropWizardDotDServices = dwDirServices

	if (OverwatchConfiguration.Verbose == true) {
		log.Printf("Overwatchd Configuration\n")
		log.Printf("------------------------------\n")
		log.Printf("  RootPath ............. : %s\n", OverwatchConfiguration.RootPath)
		log.Printf("------------------------------\n")
		log.Printf("  ServerName ........... : %s\n", OverwatchConfiguration.ServerName)
		log.Printf("  BindAddress .......... : %s\n", OverwatchConfiguration.BindAddress)
		log.Printf("  BindPort ............. : %d\n", OverwatchConfiguration.BindPort)
		log.Printf("  WebSocketUrl ......... : %s\n", OverwatchConfiguration.WebSocketUrl)
		log.Printf("------------------------------\n")
		log.Printf("  TemplatePath ......... : %s\n", OverwatchConfiguration.TemplatePath)
		log.Printf("  AssetPath ............ : %s\n", OverwatchConfiguration.AssetPath)
		log.Printf("------------------------------\n")
		log.Printf("  All (Calculated) ..... : %d\n", len(OverwatchConfiguration.LogsToWatch.All()))
		log.Printf("  StandardLogs ......... : %d\n", len(OverwatchConfiguration.LogsToWatch.StandardLogs))
		log.Printf("  DropWizardServices ... : %d\n", len(OverwatchConfiguration.LogsToWatch.DropWizardServices)*3)
		log.Printf("  DropWizardDotDServices : %d\n", len(OverwatchConfiguration.LogsToWatch.DropWizardDotDServices)*3)
		log.Printf("------------------------------\n")
		log.Printf("  LogsToWatch in Detail:\n")
		for id, lfile := range OverwatchConfiguration.LogsToWatch.All() {
			log.Printf("  %03d %s", id+1, lfile.Path)
		}
	}
}


func init() {
	flag.StringVar(&OverwatchConfiguration.ServerConfigFile, "config", "/path/to/config/default.json", "/path/to/config")
}

func main() {
	flag.Parse()
	loadConfiguration()

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

	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
