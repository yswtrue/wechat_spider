wechat_spider

微信公众号爬虫 (基于中间人攻击的爬虫核心实现,支持批量爬取公众号所有历史文章)

常见问题FAQ

代理服务端: 

通过Man-In-Middle 代理方式获取微信服务端返回,自动模拟请求自动分页,抓取对应点击的所有历史文章

客户端:  

win,macos,android,iPhone等客户端平台

代理协议:

http && https,  https需要导入certs文件夹的goproxy证书,并且添加受信权限

代理服务端

- 一个简单的Demo  simple_server.go

    package main
    
    import (
    	"log"
    	"net/http"
    
    	"github.com/sundy-li/wechat_spider"
    
    	"github.com/elazarl/goproxy"
    )
    
    func main() {
    	var port = "8899"
    	proxy := goproxy.NewProxyHttpServer()
    	//open it see detail logs
    	// wechat_spider.Verbose = true
    	proxy.OnResponse().DoFunc(
    		wechat_spider.ProxyHandle(wechat_spider.NewBaseProcessor()),
    	)
    	log.Println("server will at port:" + port)
    	log.Fatal(http.ListenAndServe(":"+port, proxy))
    
    }

- 上面贴的是一个精简的服务端,拦截客户端请求,将微信文章url打印到终端,如果想自定义输出源,可以实现Processor接口的Output方法,参考  custom_output_server.go

[1]: https://github.com/sundy-li/wechat_spider/blob/master/examples/simple_server.go
[2]: https://github.com/sundy-li/wechat_spider/blob/master/examples/custom_output_server.go
[3]: https://github.com/sundy-li/wechat_spider/blob/master/FAQ.md

- 微信会屏蔽频繁的请求,所以历史文章的翻页请求调用了Sleep()方法, 默认每个请求休眠50ms,可以根据实际情况自定义Processor覆盖此方法

客户端使用:

  (确保客户端 能正常访问 代理服务端的服务) 

- Android客户端使用方法:
  运行后, 设置手机的代理为 本机ip 8899端口,  打开微信客户端, 点击任一公众号查看历史文章按钮, 即可爬取该公众号的所有历史文章(已经支持自动翻页爬取)
- win/mac客户端,设置下全局代理对应 代理服务端的服务和端口,同理点击任一公众号查看历史文章按钮
- 自动化批量爬取所有公众号:  Windows客户端获取批量公众号所有历史文章方法,对应原理请参考 http://stackbox.cn/2016-07-21-weixin-spider-notes/ ,同时也感谢博文作者提供此windows模拟点击的思路 
  1. 要求安装windows +  微信pc版本 + ActivePython3 + autogui, 设置windows下全局代理对应 代理服务端的服务和端口
  2. 修改 win_client.py 中的 bizs参数, 通过pyautogui.position() 瞄点设置 first_ret, rel_link 坐标
  3. 在examples目录下面, 执行 python win_client.py 将自动生成链接,模拟点击


