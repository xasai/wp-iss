package main

import (
	"net"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	maxRedirects = 20
	maxBodySize  = 15000
)

var (
	paths = []string{"/", "/wordpress", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "/backup"}

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		MaxConnDuration:          time.Minute,
		MaxIdleConnDuration:      10 * time.Second,
		MaxResponseBodySize:      maxBodySize,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, 3*time.Second)
		},
	}
)

func scan(urlCh chan string, resCh chan *Resp) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer ScanWg.Done()

	bodyBuff := make([]byte, maxBodySize)

	for url := range urlCh {
		req.SetRequestURI(url)

		err := client.DoRedirects(req, res, maxRedirects)
		switch {
		case err != nil:
			continue
		case res.StatusCode() != 200:
			resCh <- &Resp{FAIL, ""}
		}

		//Save body from erase
		copy(bodyBuff, res.Body())

		if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Installation") {
			resCh <- &Resp{INSTALL, url}

		} else if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Setup Configuration File") {
			resCh <- &Resp{SETUP, url}
		} else {
			resCh <- &Resp{FAIL, url}
		}
	}
}
