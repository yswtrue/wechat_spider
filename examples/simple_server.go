package main

import (
	spider "github.com/sundy-li/wechat_spider"
)

func main() {
	var port = "8899"
	// open it see detail logs
	spider.Verbose = true
	spider.Regist(spider.NewBaseProcessor())
	spider.Run(port)
}
