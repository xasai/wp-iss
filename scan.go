package main

import (
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	maxRedirects = 20
	maxBodySize  = 15000
)

var (
	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		MaxConnDuration:          time.Minute,
		MaxIdleConnDuration:      10 * time.Second,
		MaxResponseBodySize:      maxBodySize,
	}
)

func scan(urlCh chan string, resCh chan Resp) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer ScanWg.Done()

	bodyBuff := make([]byte, maxBodySize)

	for url := range urlCh {
		req.SetRequestURI(url)

		err := client.DoRedirects(req, res, maxRedirects)
		if err != nil || res.StatusCode() != 200 {
			continue
		}
		//Save body from erase
		copy(bodyBuff, res.Body())

		if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Installation") {
			resCh <- Resp{INSTALL, url}

		} else if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Setup Configuration File") {
			resCh <- Resp{SETUP, url}
		}
	}
}
