package wechat_spider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/palantir/stacktrace"
)

type Processor interface {
	ProcessList(resp *http.Response, ctx *goproxy.ProxyCtx) ([]byte, error)
	ProcessDetail(resp *http.Response, ctx *goproxy.ProxyCtx) ([]byte, error)
	ProcessMetrics(resp *http.Response, ctx *goproxy.ProxyCtx) ([]byte, error)
	NextBiz(currentBiz string) string
	HistoryUrl() string
	Output()
	NextUrl() string
}

type BaseProcessor struct {
	req          *http.Request
	lastId       string
	data         []byte
	urlResults   []*UrlResult
	detailResult *DetailResult
	historyUrl   string
	biz          string

	// The index of urls for detail page
	currentIndex int

	Type string
}

type (
	UrlResult struct {
		Mid string
		// url
		Url  string
		_URL *url.URL
	}
	DetailResult struct {
		Url        string
		Data       []byte
		Appmsgstat *MsgStat `json:"appmsgstat"`
	}
	MsgStat struct {
		ReadNum     int `json:"read_num"`
		LikeNum     int `json:"like_num"`
		RealReadNum int `json:"real_read_num"`
	}
)

var (
	replacer = strings.NewReplacer(
		"\t", "", " ", "",
		"&quot;", `"`, "&nbsp;", "",
		`\\`, "", "&amp;amp;", "&",
		"&amp;", "&", `\`, "",
	)
	urlRegex    = regexp.MustCompile("http://mp.weixin.qq.com/s?[^#]*")
	idRegex     = regexp.MustCompile(`"id":(\d+)`)
	MsgNotFound = errors.New("MsgLists not found")

	TypeList   = "list"
	TypeDetail = "detail"
	TypeMetric = "metric"
)

func NewBaseProcessor() *BaseProcessor {
	return &BaseProcessor{}
}

func (p *BaseProcessor) init(req *http.Request, data []byte) (err error) {
	p.req = req
	p.data = data
	p.currentIndex = -1
	p.biz = req.URL.Query().Get("__biz")
	p.historyUrl = req.URL.RequestURI()
	fmt.Println("Running a new wechat processor, please wait...")
	return nil
}
func (p *BaseProcessor) ProcessList(resp *http.Response, ctx *goproxy.ProxyCtx) (data []byte, err error) {
	p.Type = TypeList
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return
	}
	if err = resp.Body.Close(); err != nil {
		return
	}

	data = buf.Bytes()
	if err = p.init(ctx.Req, data); err != nil {
		return
	}

	if err = p.processMain(); err != nil {
		return
	}

	if rootConfig.AutoScroll {
		if err = p.processPages(); err != nil {
			return
		}
	}

	cacheLock.Lock()
	defer cacheLock.Unlock()
	for _, r := range p.urlResults {
		r._URL, _ = url.Parse(r.Url)
		if r._URL != nil {
			cacheResult[r._URL.Query().Get("__biz")+"_"+r._URL.Query().Get("mid")] = &DetailResult{
				Url:        r.Url,
				Appmsgstat: &MsgStat{},
			}
		}
	}
	return
}

func (p *BaseProcessor) ProcessDetail(resp *http.Response, ctx *goproxy.ProxyCtx) (data []byte, err error) {
	p.Type = TypeDetail
	p.req = ctx.Req
	p.currentIndex++
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return
	}
	if err = resp.Body.Close(); err != nil {
		return
	}
	data = buf.Bytes()

	result := cacheResult[genKey(p.req.URL)]
	if result == nil {
		result = &DetailResult{}
		cacheLock.Lock()
		defer cacheLock.Unlock()
		cacheResult[genKey(p.req.URL)] = result
	}
	result.Data = data
	return
}

func (p *BaseProcessor) ProcessMetrics(resp *http.Response, ctx *goproxy.ProxyCtx) (data []byte, err error) {
	p.Type = TypeMetric
	p.req = ctx.Req

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return
	}
	if err = resp.Body.Close(); err != nil {
		return
	}
	data = buf.Bytes()
	detailResult := &DetailResult{}
	e := json.Unmarshal(data, detailResult)
	if e != nil {
		p.logf("error in parsing json %s\n", string(data))
	}
	//must be not nil
	result := cacheResult[genKey(p.req.URL)]
	if result == nil {
		result = &DetailResult{}
	}
	result.Appmsgstat = detailResult.Appmsgstat
	p.detailResult = result
	return
}

func (p *BaseProcessor) NextBiz(currentBiz string) string {
	return ""
}

func (p *BaseProcessor) NextUrl() string {
	if p.currentIndex+1 < len(p.urlResults) {
		return p.urlResults[p.currentIndex+1].Url + "&ttt=1111"
	}
	return ""
}

func (p *BaseProcessor) HistoryUrl() string {
	return p.historyUrl
}

func (p *BaseProcessor) Sleep() {
	time.Sleep(50 * time.Millisecond)
}

func (p *BaseProcessor) UrlResults() []*UrlResult {
	return p.urlResults
}

func (p *BaseProcessor) DetailResult() *DetailResult {
	return p.detailResult
}

func (p *BaseProcessor) GetRequest() *http.Request {
	return p.req
}

func (p *BaseProcessor) Output() {
	urls := []string{}
	fmt.Println("result => [")
	for _, r := range p.urlResults {
		urls = append(urls, r.Url)
	}
	fmt.Println(strings.Join(urls, ","))
	fmt.Println("]")
}

//Parse the html
func (p *BaseProcessor) processMain() error {
	p.urlResults = make([]*UrlResult, 0, 100)
	buffer := bytes.NewBuffer(p.data)
	var msgs string
	str, err := buffer.ReadString('\n')
	for err == nil {
		if strings.Contains(str, "msgList = ") {
			msgs = str
			break
		}
		str, err = buffer.ReadString('\n')
	}
	if msgs == "" {
		return stacktrace.Propagate(MsgNotFound, "Failed parse main")
	}
	msgs = replacer.Replace(msgs)
	urls := urlRegex.FindAllString(msgs, -1)
	if len(urls) < 1 {
		return stacktrace.Propagate(MsgNotFound, "Failed find url in  main")
	}
	p.urlResults = make([]*UrlResult, len(urls))
	for i, u := range urls {
		p.urlResults[i] = &UrlResult{Url: u}
	}

	idMatcher := idRegex.FindAllStringSubmatch(msgs, -1)
	if len(idMatcher) < 1 {
		return stacktrace.Propagate(MsgNotFound, "Failed find id in  main")
	}
	p.lastId = idMatcher[len(idMatcher)-1][1]
	return nil
}

func (p *BaseProcessor) processPages() (err error) {
	var pageUrl = p.genPageUrl()
	p.logf("process pages....")
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return stacktrace.Propagate(err, "Failed new page request")
	}
	for k, _ := range p.req.Header {
		req.Header.Set(k, p.req.Header.Get(k))
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stacktrace.Propagate(err, "Failed get page response")
	}
	bs, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	str := replacer.Replace(string(bs))
	result := urlRegex.FindAllString(str, -1)
	if len(result) < 1 {
		return stacktrace.Propagate(err, "Failed get page url")
	}
	idMatcher := idRegex.FindAllStringSubmatch(str, -1)
	if len(idMatcher) < 1 {
		return stacktrace.Propagate(err, "Failed get page id")
	}
	p.lastId = idMatcher[len(idMatcher)-1][1]
	p.logf("Page Get => %d,lastid: %s", len(result), p.lastId)
	for _, u := range result {
		p.urlResults = append(p.urlResults, &UrlResult{Url: u})
	}
	if p.lastId != "" {
		p.Sleep()
		return p.processPages()
	}
	return nil
}

func (p *BaseProcessor) genPageUrl() string {
	urlStr := "http://mp.weixin.qq.com/mp/getmasssendmsg?" + p.req.URL.RawQuery
	urlStr += "&frommsgid=" + p.lastId + "&f=json&count=100"
	return urlStr
}

func genKey(uri *url.URL) string {
	return uri.Query().Get("__biz") + "_" + uri.Query().Get("mid")
}

func (P *BaseProcessor) logf(format string, msg ...interface{}) {
	if rootConfig.Verbose {
		Logger.Printf(format, msg...)
	}
}
