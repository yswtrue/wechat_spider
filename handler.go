package wechat_spider

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/elazarl/goproxy"
)

var (
	Verbose = false
	Logger  = log.New(os.Stderr, "", log.LstdFlags)
)

func ProxyHandle(proc Processor) func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	return func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		Logger.Println("Hijacked of", ctx.Req.URL.RequestURI())
		if ctx.Req.URL.Path == `/mp/getmasssendmsg` && !strings.Contains(ctx.Req.URL.RawQuery, `f=json`) {
			var data []byte
			var err error
			data, resp.Body, err = copyReader(resp.Body)
			if err != nil {
				return resp
			}
			t := reflect.TypeOf(proc)
			v := reflect.New(t.Elem())
			p := v.Interface().(Processor)
			go func() {
				err = p.Process(ctx.Req, data)
				if err != nil {
					Logger.Println(err.Error())
				}
				p.Output()
			}()
		}
		return resp
	}

}

// One of the copies, say from b to r2, could be avoided by using a more
func copyReader(b io.ReadCloser) (bs []byte, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
