package wechat_spider

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"

	"github.com/elazarl/goproxy"
)

var (
	Logger      = log.New(os.Stderr, "", log.LstdFlags)
	procs       = make(map[string]Processor, 10)
	cacheResult = make(map[string]*DetailResult)
	cacheLock   sync.Mutex
)

func ProxyHandle(proc Processor) func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp == nil || resp.StatusCode != 200 {
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
	go p.Output()
	var nextUrl = p.NextUrl()
	if !needDetail || nextUrl == "" {
		curBiz := ctx.Req.URL.Query().Get("__biz")
		nextBiz := p.NextBiz(curBiz)
		if nextBiz != "" {
			nextUrl = fmt.Sprintf("http://mp.weixin.qq.com/mp/getmasssendmsg?__biz=%s#wechat_webview_type=1&wechat_redirect", nextBiz)
		}
	}
	var buf = bytes.NewBuffer(data)

	if nextUrl != "" {
		println("nexturl list==>", nextUrl)
		buf.WriteString(fmt.Sprintf(`<script>setTimeout(function(){window.location.href="%s";},2000);</script>`, nextUrl))
	}
	saveProcessor(ctx.Req, p)
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
			nextUrl = fmt.Sprintf("http://mp.weixin.qq.com/mp/getmasssendmsg?__biz=%s#wechat_webview_type=1&wechat_redirect", nextBiz)
		}
	}
	var buf = bytes.NewBuffer(data)

	if nextUrl != "" {
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
	saveProcessor(ctx.Req, p)
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
}

// get the processor from cache
func getProcessor(req *http.Request, proc Processor) Processor {
	key := hashKey(req.Header.Get("q-guid"))

	if p, ok := procs[key]; ok {
		return p
	}

	t := reflect.TypeOf(proc)
	v := reflect.New(t.Elem())
	p := v.Interface().(Processor)
	return p
}

func saveProcessor(req *http.Request, proc Processor) {
	key := hashKey(req.Header.Get("q-guid"))
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if proc == nil {
		delete(procs, key)
		return
	}
	procs[key] = proc
}

func hashKey(key string) string {
	h := md5.New()
	io.WriteString(h, key)
	return fmt.Sprintf("%x", h.Sum(nil))
}
