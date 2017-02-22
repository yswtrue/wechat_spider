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
	case spider.TypeList:
		//do nothing
		fmt.Printf("url size ==> %#v\n", len(c.UrlResults()))
	case spider.TypeDetail:
		fmt.Printf("url %s %s is being spidered\n", c.DetailResult().Id, c.DetailResult().Url)
	case spider.TypeMetric:
		fmt.Printf("url %s %s metric %#v is being spidered\n", c.DetailResult().Id, c.DetailResult().Url, c.DetailResult().Appmsgstat)
	}
}

// NextBiz hijack the script, set the location to next url after 2 seconds
func (c *CustomProcessor) NextBiz(currentBiz string) string {
	// Random select
	return _bizs[rand.Intn(len(_bizs))]
}

// NextUrl hijack the script, set the location to next url after 2 seconds
func (c *CustomProcessor) NextUrl(currentUrl string) string {
	// Random select
	return _urls[rand.Intn(len(_urls))]
}

var (
	_bizs = []string{"MzAwODI2OTA1MA==", "MzA5NDk4ODI4Mw==", "MjM5MjEyOTEyMQ=="}
	_urls = []string{
		"http://mp.weixin.qq.com/s?__biz=MzI2MzMxNzEzNA==&mid=2247484076&idx=1&sn=2b4b1dd2001d525e08966be9198d3f8d&scene=2&srcid=0804vRsESdgOtdaWTx4CET9Y&from=timeline&isappinstalled=0#wechat_redirect",

		"http://mp.weixin.qq.com/s?__biz=MjM5MDMyMzg2MA==&mid=2655481188&idx=1&sn=b22c1b7089ef132724d5c35b82372cb3&chksm=bdf5489f8a82c1895c72f488602a1f09b2ba7821a61d2499bf6ad35c88dcfa737af6675065a7#rd",

		"http://mp.weixin.qq.com/s?__biz=MjM5MDMyMzg2MA==&mid=2655481311&idx=4&sn=96a00896b5c800325ba5fc978ec15614&chksm=bdf548248a82c13223989f9469219a040d53c145eb9db4c309a6f4dfb1e4ed1a76dcc22b7196#rd",
	}
)
