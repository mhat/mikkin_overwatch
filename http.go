package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"io"
	"net/http"
        "github.com/hoisie/mustache"
	"encoding/json"
)

func SockServer(ws *websocket.Conn) {
	var err error
	var clientMessage string

	// cleanup on server side
	defer func() {
		if err = ws.Close(); err != nil {
			log.Println("Websocket could not be closed", err.Error())
		}
	}()

	cc := ClientConnection{websocket: ws, clientIP: ws.Request().RemoteAddr}
	ClientConnections[cc] = 0

	for {
		if err = websocket.JSON.Receive(ws, &clientMessage); err != nil {
			// If we cannot Read then the connection is closed
			log.Println("Websocket Disconnected waiting", err.Error())
			delete(ClientConnections, cc)
			return
		}
		log.Println("Client [" + cc.clientIP + "] Sent: " + clientMessage)
	}
}

func LogsHandler(w http.ResponseWriter, req *http.Request) {
	b, _ := json.Marshal(MikkinStreamFiles)
	w.Write(b)
}

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Contet-Type", "text/html")
	io.WriteString(w, mustache.RenderFile("templates/console.html.mustache"))
}

