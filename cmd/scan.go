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
	defer scanWg.Done()
	for url := range urlCh {

		req.SetRequestURI(url)

		status, err := scanUnit(req, res)

		if err != nil || res.StatusCode() != fasthttp.StatusOK {
			resCh <- &Resp{FAIL, ""}
			continue
		}

		// Go to next url if I/S found
		if status != FAIL {
			resCh <- &Resp{status, url}
			continue
		}

		// Scan every sub path in url
		for _, path := range paths {
			req.SetRequestURI(url + path)
			status, _ = scanUnit(req, res)
			// break path loop if succeed
			if status != FAIL {
				resCh <- &Resp{status, url + path}
				break
			}
		}

		if status == FAIL {
			resCh <- &Resp{FAIL, url}
		}
	}
}

func scanUnit(req *fasthttp.Request, res *fasthttp.Response) (Status, error) {

	atomic.AddUint64(&reqTotal, 1)

	err := client.DoRedirects(req, res, maxRedirects)
	if err != nil {
		errlog.Println(err.Error(), res.StatusCode(), req.URI())
		return FAIL, err
	}
	if res.StatusCode() != fasthttp.StatusOK {
		return FAIL, nil
	}
	return checkResponse(res.String()), nil
}

func checkResponse(body string) Status {
	if strings.Contains(body, "WordPress &rsaquo; Installation") {
		return INSTALL
	}
	if strings.Contains(body, "WordPress &rsaquo; Setup Configuration File") {
		return SETUP
	}
	return FAIL
}
