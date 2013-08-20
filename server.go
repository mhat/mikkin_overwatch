package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"io"
	"net/http"
        "github.com/hoisie/mustache"
	"encoding/json"
	"path"
)

var ClientConnections = make(map[string]ClientConnection)
var ChannelToWS = make(chan WSServerMessage)

type ClientConnection struct {
	websocket *websocket.Conn
	clientIP  string
	key string
	filters map[string]int
}

func NewClientConnection (ws *websocket.Conn) ClientConnection {
	return ClientConnection{
		websocket: ws,
		key: ws.Request().Header["Sec-Websocket-Key"][0],
		clientIP:  ws.Request().RemoteAddr,
		filters: make(map[string]int)}
}


type WSClientMessage struct {
	Channel string
	Message string
}


type WSServerMessage struct {
	Channel string
	Content string
}

func SockServer(ws *websocket.Conn) {
	var err error
	var clientMessage WSClientMessage

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	cc := NewClientConnection(ws)
	ClientConnections[cc.key] = cc
	log.Printf("Websocket Connected Client [%s].\n", cc.clientIP)

	// update the client with state!
	for i, _ := range MikkinStreamFiles {
		msf := &MikkinStreamFiles[i]
		msf.buffer.Do(func(p interface{}) {
			if (p != nil) {
				websocket.JSON.Send(ws, WSServerMessage{msf.Path, p.(string)})
			}
		})
	}

	for {
		if err = websocket.JSON.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			delete(ClientConnections, cc.key)
			return
		}

		switch clientMessage.Message {
		case "subscribe":
			//log.Printf("Recevied subscribe from client %s, msg=[%s]", cc.clientIP, clientMessage)
			delete(cc.filters, clientMessage.Channel)
		case "unsubscribe":
			//log.Printf("Recevied unsubscribe from client %s, msg=[%s]", cc.clientIP, clientMessage)
			cc.filters[clientMessage.Channel] = 1
		default:
			log.Printf("Recevied unknown message type from client %s, msg=[%s]", cc.clientIP, clientMessage)
		}
	}
}


func LogsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Contet-Type", "text/json")
	b, _ := json.Marshal(MikkinStreamFiles)
	w.Write(b)
}


type ConsoleView struct {
	WebSocketUrl string
}


func HomeHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := path.Join(OverwatchConfiguration.TemplatePath, "console.html.mustache")
	view := ConsoleView{WebSocketUrl: OverwatchConfiguration.WebSocketUrl.String()}
	w.Header().Set("Contet-Type", "text/html")
	io.WriteString(w, mustache.RenderFile(tmpl, view))
}




