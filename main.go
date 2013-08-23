package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"fmt"
	"net/http"
	"flag"
)

func init() {
	flag.StringVar(&OverwatchConfiguration.ServerConfigFile, "config", "config/server.json", "/path/to/config")
}


func main() {
	flag.Parse()
	OverwatchConfiguration.LoadConfiguration()

	// create connection manager
	cm := NewConnectionManager()

	// watch some files
	go FileWatcher(OverwatchConfiguration.LogsToWatch.All(), cm.BroadcastChannel)
	go MessageBroadcaster(cm)

	// cute
	go BingImageOfTheDayPoller()

	log.Printf("------------------------------\n")
	log.Printf("Register Static Assets\n")
	http.Handle("/assets/",   http.FileServer(http.Dir(OverwatchConfiguration.RootPath)))

	log.Printf("Registering Views\n")
	http.HandleFunc("/",      HomeHandler)

	log.Printf("Registering Resource\n")
	http.HandleFunc("/logs",  LogsHandler)

	log.Printf("Register WebSocket Handler\n")
	http.Handle("/websocket", websocket.Handler(NewWebsocketHandler(cm)))
	log.Printf("------------------------------\n")

        bindAddressAndPort := fmt.Sprintf("%s:%d", OverwatchConfiguration.BindAddress, OverwatchConfiguration.BindPort)
        cert := "/etc/ssl/certs/any.yammer.dev.pem"
        key  := "/etc/ssl/private/any.yammer.dev.key"
        err := http.ListenAndServeTLS(bindAddressAndPort, cert, key, nil)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

}
