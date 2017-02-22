package main

import (
	"fmt"
	"math/rand"

	spider "github.com/sundy-li/wechat_spider"
)

func main() {
	var port = "8899"
	spider.Regist(&CustomProcessor{})
	spider.Run(port)

}

//Just to implement Output Method of interface{} Processor
type CustomProcessor struct {
	spider.BaseProcessor
}

func (c *CustomProcessor) Output() {
	urls := []string{}
	for _, r := range c.UrlResults() {
		urls = append(urls, r.Url)
	}
	fmt.Printf("%#v\n", urls)
	// You can dump the get the html from urls and save to your database
}

// NextBiz hijack the script, set the location to next url after 2 seconds
func (c *CustomProcessor) NextBiz(currentBiz string) string {
	// Random select
	return _bizs[rand.Intn(len(_bizs))]
}

var (
	_bizs = []string{"MzAwODI2OTA1MA==", "MzA5NDk4ODI4Mw==", "MjM5MjEyOTEyMQ=="}
)
