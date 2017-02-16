package main

import (
	"fmt"
	"math/rand"

	spider "github.com/sundy-li/wechat_spider"
)

func main() {
	var port = "8899"
	spider.InitConfig(&spider.Config{
		Verbose:    false, // Open to see detail logs
		AutoScroll: false, // Open to crawl scroll pages
		Metrics:    true,
	})
	spider.Regist(&CustomProcessor{})
	spider.Run(port)
}

//Just to implement Output Method of interface{} Processor
type CustomProcessor struct {
	spider.BaseProcessor
}

func (c *CustomProcessor) Output() {
	switch c.Type {
	case spider.TypeDetail:
		fmt.Printf("url %s size ==> %#v\n", c.DetailResult().Url, len(c.DetailResult().Data))
	case spider.TypeMetric:
		fmt.Printf("%#v\n", c.DetailResult().Appmsgstat)
	}
}

// NextBiz hijack the script, set the location to next url after 2 seconds
func (c *CustomProcessor) NextBiz(currentBiz string) string {
	// Random select
	return _bizs[rand.Intn(len(_bizs))]
}

var (
	_bizs = []string{"MzAwODI2OTA1MA==", "MzA5NDk4ODI4Mw==", "MjM5MjEyOTEyMQ=="}
)
