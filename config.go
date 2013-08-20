package main

import (
	"log"
	"fmt"
	"os"
	"net/url"
	"encoding/json"
	"io/ioutil"
 )


var OverwatchConfiguration OverwatchConfigurationStruct


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

func (*OverwatchConfigurationStruct) LoadConfiguration() {
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

	uniqueDropWizardServices := map[string]bool{}

	for _, svc := range logs.DropWizardServices {
		uniqueDropWizardServices[svc] = true
	}

	for _, svc := range logs.DropWizardDotDServices {
		uniqueDropWizardServices[svc] = true
	}

	for svc, _ := range uniqueDropWizardServices {
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




