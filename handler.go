package wechat_spider

import (
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/elazarl/goproxy"
)

var (
	Verbose = false
	Logger  = log.New(os.Stderr, "", log.LstdFlags)
)

func ProxyHandle(proc Processor) func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		if resp.StatusCode != 200 {
			return resp
		}
		if Verbose {
			Logger.Println("Hijacked of", ctx.Req.URL.RequestURI())
		}
		if ctx.Req.URL.Path == `/mp/getmasssendmsg` || (ctx.Req.URL.Path == `/mp/profile_ext` && ctx.Req.URL.Query().Get("action") == "home") {
			//&& !strings.Contains(ctx.Req.URL.RawQuery, `f=json`)
			var err error
			t := reflect.TypeOf(proc)
			v := reflect.New(t.Elem())
			p := v.Interface().(Processor)
			err = p.Process(resp, ctx)
			if err != nil {
				Logger.Println(err.Error())
			}
			go func() {
				p.Output()
			}()
		}
		return resp
	}

}
