package main

import (
	"net/http"
	"net/url"
	"launchpad.net/xmlpath"
	"time"
	"fmt"
	"log"
)

var BingImageOfTheDayUrl = ""

func BingImageOfTheDayPoller () {
	bingImageApiUrl := "http://www.bing.com/HPImageArchive.aspx?format=xml&idx=0&n=1&mkt=en-US"
	duration, _ := time.ParseDuration("1h")
	for {
		resp, _ := http.Get(bingImageApiUrl)
		path    := xmlpath.MustCompile("/images/image/url")
		root, _ := xmlpath.Parse(resp.Body)
		if text, ok := path.String(root); ok {
			BingImageOfTheDayUrl = fmt.Sprintf("http://www.bing.com%s", text)
			log.Printf("Setting BingImageOfTheDay: %s\n", BingImageOfTheDayUrl)
		}

		time.Sleep(duration)
	}
}

func GetBingImageOfTheDayUrl() string {
	// fallback
	if BingImageOfTheDayUrl == "" {
		imageUrl, _ := url.Parse(fmt.Sprintf("http://%s:%d/%s",
			OverwatchConfiguration.ServerName,
			OverwatchConfiguration.BindPort,
			"assets/img/bg1.jpg"))
		BingImageOfTheDayUrl = imageUrl.String()
	}
	return BingImageOfTheDayUrl
}
