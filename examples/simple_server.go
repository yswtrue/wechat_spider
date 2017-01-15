package main

import (
	spider "github.com/sundy-li/wechat_spider"
)

func main() {
	var port = "8899"
	spider.InitConfig(&spider.Config{
		Verbose:    false, // Open to see detail logs
		AutoScroll: false, // Open to crawl scroll pages
	})
	spider.Regist(spider.NewBaseProcessor())
	spider.Run(port)
}
