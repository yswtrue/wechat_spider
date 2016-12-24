package main

import (
	spider "github.com/sundy-li/wechat_spider"
)

func main() {
	var port = "8899"
	// open it see detail logs
	spider.Verbose = true
	spider.Regist(&CustomProcessor{})
	spider.Run(port)

}

//Just to implement Output Method of interface{} Processor
type CustomProcessor struct {
	spider.BaseProcessor
}

func (c *CustomProcessor) Output() {
	// Just print the length of result urls
	println("result urls size =>", len(c.Urls()))
	// You can dump the get the html from urls and save to your database
}
