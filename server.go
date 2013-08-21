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



type WSClientMessage struct {
	Channel string
	Message string
}

type WSServerMessage struct {
	Channel string
	Content string
}



type ClientConnection struct {
	websocket *websocket.Conn
	clientAddress string
	key string
	filters map[string]int
}

func NewClientConnection (ws *websocket.Conn) ClientConnection {
	return ClientConnection{
		websocket: ws,
		key: ws.Request().Header["Sec-Websocket-Key"][0],
		clientAddress:  ws.Request().RemoteAddr,
		filters: make(map[string]int)}
}



type ConnectionManager struct {
	connections map[string]ClientConnection
	BroadcastChannel chan WSServerMessage
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]ClientConnection),
		BroadcastChannel: make(chan WSServerMessage)}
}

func (cm *ConnectionManager) RegisterClientConnection (conn ClientConnection) {
	log.Printf("Registered client %s\n", conn.key)
	cm.connections[conn.key] = conn
}

func (cm *ConnectionManager) RemoveClientConnection (conn ClientConnection) {
	log.Printf("Removing client %s\n", conn.key)
	delete(cm.connections, conn.key)
}

func (cm *ConnectionManager) ClientsCount () int {
	return len(cm.connections)
}



func MessageBroadcaster(cm *ConnectionManager) {
	for {
		// blocking read: waiting for messages on this channel
		msg := <-cm.BroadcastChannel
		// log.Printf("BroadcastChannel [%s]\n", msg)
		for _, client := range cm.connections {
			_, mf := client.filters[msg.Channel]
			if (mf == true) {
				continue
			}

			if err := websocket.JSON.Send(client.websocket, msg); err != nil {
				log.Println("Error sending to client %s. %s\n", client.clientAddress, err.Error())
			}
		}
	}

}



func NewWebsocketHandler (cm *ConnectionManager) func(*websocket.Conn) {
	return func(ws *websocket.Conn){
		websocketHandler(cm, ws)
	}
}

func websocketHandler(cm *ConnectionManager, ws *websocket.Conn) {
	var err error
	var clientMessage WSClientMessage

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	client := NewClientConnection(ws)
	cm.RegisterClientConnection(client)

	// update the client with state!
	for i, _ := range WatchedLogFiles {
		wlog := &WatchedLogFiles[i]
		wlog.Buffer.Do(func(p interface{}) {
			if (p != nil) {
				websocket.JSON.Send(ws, WSServerMessage{wlog.Info.Path, p.(string)})
			}
		})
	}

	for {
		if err = websocket.JSON.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			cm.RemoveClientConnection(client)
			return
		}

		switch clientMessage.Message {
		case "subscribe":
			//log.Printf("Recevied subscribe from client %s, msg=[%s]", client.clientAddress, clientMessage)
			delete(client.filters, clientMessage.Channel)
		case "unsubscribe":
			//log.Printf("Recevied unsubscribe from client %s, msg=[%s]", client.clientAddress, clientMessage)
			client.filters[clientMessage.Channel] = 1
		default:
			log.Printf("Recevied unknown message type from client %s, msg=[%s]", client.clientAddress, clientMessage)
		}
	}
}



func LogsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Contet-Type", "text/json")
	b, _ := json.Marshal(WatchedLogFiles)
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




