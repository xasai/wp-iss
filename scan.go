package main

import (
	"strings"
	"sync/atomic"

	"github.com/valyala/fasthttp"
)

var paths = []string{"/wp-admin", "/wordpress", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "/backup"}

func scan(urlCh chan string, resCh chan *Resp) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer ScanWg.Done()

	//bodyBuff := make([]byte, maxBodySize)

	for url := range urlCh {
		req.SetRequestURI(url)
		req.SetConnectionClose()

		atomic.AddUint64(&reqTotal, 1)

		err := client.DoRedirects(req, res, maxRedirects)
		if err != nil || res.StatusCode() != fasthttp.StatusOK {
			resCh <- &Resp{FAIL, ""}
			continue
		} /* else if res.StatusCode() != fasthttp.StatusOK { // code != 200
			resCh <- &Resp{FAIL, ""}
			continue
		}*/

		if strings.Contains(res.String(), "WordPress &rsaquo; Installation") {
			resCh <- &Resp{INSTALL, url}
			logger.Println(url, res.String())
		} else if strings.Contains(res.String(), "WordPress &rsaquo; Setup Configuration File") {
			resCh <- &Resp{SETUP, url}
			logger.Println(url, res.String())
		} else {
			resCh <- &Resp{FAIL, url}
		}
	}
}

func scan2(urlCh chan string, resCh chan *Resp) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer ScanWg.Done()

	for url := range urlCh {
		req.SetRequestURI(url)

		atomic.AddUint64(&reqTotal, 1)
		err := client.DoRedirects(req, res, maxRedirects)
		if err != nil || res.StatusCode() != fasthttp.StatusOK {
			resCh <- &Resp{FAIL, ""}
			continue
		}
		if strings.Contains(res.String(), "WordPress &rsaquo; Installation") {
			resCh <- &Resp{INSTALL, url}
			logger.Println(url, res.String())
		}
		if strings.Contains(res.String(), "WordPress &rsaquo; Setup Configuration File") {
			resCh <- &Resp{SETUP, url}
			logger.Println(url, res.String())
		}
		for path := range paths {

		}
	}
}
