package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

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
	// You can write the result to files
	f, _ := os.OpenFile("result.jsons", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	for _, u := range c.Urls() {
		resp, err := http.Get(u)
		if err != nil {
			println(err.Error())
			continue
		}
		bs, _ := ioutil.ReadAll(resp.Body)
		mp := map[string]interface{}{
			"url":  u,
			"data": string(bs),
		}
		bs, _ = json.Marshal(mp)
		f.Write(bs)
		f.WriteString("\n")
	}
}
