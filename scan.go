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

func scan(urlCh, insCh, setCh, failCh chan string) {

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
			failCh <- url
			continue
		}

		//Save body from erase
		copy(bodyBuff, res.Body())

		if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Installation") {
			insCh <- url
		} else if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Setup Configuration File") {
			setCh <- url
		} else {
			failCh <- url
		}
	}
}
