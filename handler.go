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
	Logger = log.New(os.Stderr, "", log.LstdFlags)
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
			//&& !strings.Contains(ctx.Req.URL.RawQuery, `f=json`)
			var err error
			t := reflect.TypeOf(proc)
			v := reflect.New(t.Elem())
			p := v.Interface().(Processor)
			data, err := p.Process(resp, ctx)
			if err != nil {
				Logger.Println(err.Error())
			}

			var buf = bytes.NewBuffer(data)
			//Auto location
			curBiz := ctx.Req.URL.Query().Get("__biz")
			nextBiz := p.NextBiz(curBiz)
			if nextBiz != "" {
				nextUrl := strings.Replace(ctx.Req.URL.RequestURI(), curBiz, nextBiz, -1)
				buf.WriteString(fmt.Sprintf(`<script>setTimeout(function(){window.location.href="%s";},2000);</script>`, nextUrl))
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

			go func() {
				p.Output()
			}()
		}
		return resp
	}

}
