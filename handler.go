package wechat_spider

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/elazarl/goproxy"
)

var (
	Logger      = log.New(os.Stderr, "", log.LstdFlags)
	procs       = make(map[string]Processor, 10)
	cacheResult = make(map[string]*DetailResult)
)

func ProxyHandle(proc Processor) func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.StatusCode != 200 {
			return resp
		}
		if rootConfig.Verbose {
			Logger.Println("Hijacked of", ctx.Req.URL.RequestURI())
		}

		if ctx.Req.URL.Path == `/mp/getmasssendmsg` || (ctx.Req.URL.Path == `/mp/profile_ext` && ctx.Req.URL.Query().Get("action") == "home") {
			handleList(resp, ctx, proc)
		} else if ctx.Req.URL.Path == `/s` {
			handleDetail(resp, ctx, proc)
		} else if ctx.Req.URL.Path == "/mp/getappmsgext" {
			handleMetrics(resp, ctx, proc)
		}
		return resp
	}

}

func handleList(resp *http.Response, ctx *goproxy.ProxyCtx, proc Processor) {
	needDetail := rootConfig.Metrics
	var err error
	p := getProcessor(ctx.Req, proc)
	data, err := p.ProcessList(resp, ctx)
	if err != nil {
		log.Println(err.Error())
	}
	// When fetch metrics, list page output could be ingore
	var nextUrl = p.NextUrl()
	if !needDetail || nextUrl == "" {
		go p.Output()
		curBiz := ctx.Req.URL.Query().Get("__biz")
		nextBiz := p.NextBiz(curBiz)
		if nextBiz != "" {
			nextUrl = strings.Replace(p.HistoryUrl(), curBiz, nextBiz, -1)
		}
	}
	var buf = bytes.NewBuffer(data)

	if nextUrl != "" {
		println("nexturl list==>", nextUrl)
		buf.WriteString(fmt.Sprintf(`<script>setTimeout(function(){window.location.href="%s";},2000);</script>`, nextUrl))
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
}

func handleDetail(resp *http.Response, ctx *goproxy.ProxyCtx, proc Processor) {
	needDetail := rootConfig.Metrics
	var err error
	p := getProcessor(ctx.Req, proc)
	data, err := p.ProcessDetail(resp, ctx)
	if err != nil {
		log.Println(err.Error())
	}
	// When fetch metrics, list page output could be ingore
	var nextUrl = p.NextUrl()
	if !needDetail || nextUrl == "" {
		curBiz := ctx.Req.URL.Query().Get("__biz")
		nextBiz := p.NextBiz(curBiz)
		if nextBiz != "" {
			nextUrl = strings.Replace(p.HistoryUrl(), curBiz, nextBiz, -1)
		}
	}
	var buf = bytes.NewBuffer(data)

	if nextUrl != "" {
		println("nexturl detail==>", nextUrl)
		buf.WriteString(fmt.Sprintf(`<script>setTimeout(function(){window.location.href="%s";},2000);</script>`, nextUrl))
	}
	go p.Output()
	saveProcessor(ctx.Req, p)
	resp.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
}

func handleMetrics(resp *http.Response, ctx *goproxy.ProxyCtx, proc Processor) {
	var err error
	p := getProcessor(ctx.Req, proc)
	data, err := p.ProcessMetrics(resp, ctx)
	if err != nil {
		Logger.Println(err.Error())
	}
	go p.Output()
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
}

// get the processor from cache
func getProcessor(req *http.Request, proc Processor) Processor {
	key := req.Header.Get("q-guid")
	if p, ok := procs[key]; ok {
		return p
	}

	t := reflect.TypeOf(proc)
	v := reflect.New(t.Elem())
	p := v.Interface().(Processor)
	//set
	procs[key] = p

	return p
}

func saveProcessor(req *http.Request, proc Processor) {
	key := req.Header.Get("x-wechat-uin") + req.UserAgent()
	procs[key] = proc
}
