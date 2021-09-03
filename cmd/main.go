package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/schollz/progressbar/v3"
	"github.com/valyala/fasthttp"
)

const (
	banner = `
	 _    _ ______       _____  _____  _____ 
	| |  | || ___ \     |_   _|/  ___|/  ___|
	| |  | || |_/ /______ | |  \ '--. \ '--. 
	| |/\| ||  __/|______|| |   '--. \ '--. \
	\  /\  /| |          _| |_ /\__/ //\__/ /
	 \/  \/ \_|          \___/ \____/ \____/ 
												
	     WP Install and Setup Scanner
`
	usage = "USAGE:  [OPTION] domain_list"

	maxRedirects = 20

	maxBodySize = 300000
	minBodySize = 2000
)

var (
	//Flags
	jobs        int   = 100
	dialTimeout int64 = 3

	reqTotal uint64
	logger   *log.Logger
	errlog   *log.Logger
	client   *fasthttp.Client
	scanWg   sync.WaitGroup
	respWg   sync.WaitGroup
)

func main() {
	fmt.Println(banner)
	fmt.Println(" jobs:", jobs)

	// Read domain list from specefied file
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	d, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	URLs := strings.Split(strings.Replace(string(d), "\r\n", "\n", -1), "\n")
	URLs = URLs[:len(URLs)-1]

	// Create file to write install result
	installFile, err := os.OpenFile("result/install.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer installFile.Close()

	// Create file to write setup result
	setupFile, err := os.OpenFile("result/setup.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer setupFile.Close()

	//main will send urls from domain list for scanner routines
	urlCh := make(chan string)

	//printer routine will receive result from scanner routines
	resCh := make(chan *Resp, 10)

	//launch scanners
	scanWg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go scan(urlCh, resCh)
	}

	//launch printer routine

	go func() {
		respWg.Add(1)
		defer respWg.Done()

		bar := progressbar.NewOptions(
			len(URLs),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionFullWidth(),
		)
		tick := time.Tick(time.Second)
		go func() {
			for {
				<-tick
				bar.Add(0)
			}
		}()
		defer bar.Close()
		for resp := range resCh {
			switch resp.Status {
			case INSTALL:
				bar.Clear()
				fmt.Printf(" [+] Install ===> %s\n", resp.Url)
				bar.Add(1)
				logger.Printf(" Install %s\n", resp.Url)
				installFile.WriteString(resp.Url + "\n")
			case SETUP:
				bar.Clear()
				fmt.Printf(" [+] Setup ===> %s\n", resp.Url)
				bar.Add(1)
				logger.Printf(" Setup %s\n", resp.Url)
				setupFile.WriteString(resp.Url + "\n")
			case FAIL:
				bar.Add(1)
			}
		}
	}()

	//timer
	start := time.Now()

	// Send http:// + url to scanners routines
	for _, URL := range URLs {
		urlCh <- "http://" + URL
	}

	//Wait scanners end
	close(urlCh)
	scanWg.Wait()

	//Wait printer end
	close(resCh)
	respWg.Wait()

	//print time scanner work
	fmt.Println(time.Now().Sub(start), "Requests total:", reqTotal)
	logger.Println(time.Now().Sub(start), "Requests total:", reqTotal)
}

func init() {

	var logOn bool

	flag.IntVar(&jobs, "j", jobs, "number of goroutines to run")
	flag.IntVar(&jobs, "jobs", jobs, "number of goroutines to run")
	flag.Int64Var(&dialTimeout, "t", dialTimeout, "dial timeout")
	flag.BoolVar(&logOn, "l", false, "enables error log and bench")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		MaxConnDuration:          20 * time.Second,
		MaxIdleConnDuration:      20 * time.Second,
		ReadTimeout:              5 * time.Second,
		WriteTimeout:             5 * time.Second,
		MaxConnWaitTimeout:       5 * time.Second,
		MaxResponseBodySize:      maxBodySize,

		//Specifying default tcp timeout diealer with dialTimout seconds
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*time.Duration(dialTimeout))
		},
	}

	os.Mkdir("result", 0777)

	if logOn {
		logInit(logOn)
		return
	}
	logger = log.New(ioutil.Discard, "", 0)
	errlog = log.New(ioutil.Discard, "", 0)
}

func logInit(logOn bool) {

	benchFilename := "result/bench_" + filepath.Base(flag.Arg(0))
	benchFile, err := os.OpenFile(benchFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	logger = log.New(benchFile, "", 0)

	errlogFilename := "result/error"
	errlogFile, err := os.OpenFile(errlogFilename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	errlog = log.New(errlogFile, "", 0)

	//Print init info in benchLog
	logger.Println(" =======================================")
	logger.Println(" jobs:", jobs,
		"| MaxConnDuration: ", client.MaxConnDuration,
		"| MaxIdleConnDuration: ", client.MaxIdleConnDuration,
		"| ReadTimeout: ", client.ReadTimeout,
		"\n WriteTimeout: ", client.WriteTimeout,
		"| MaxConnWaitTimeout: ", client.MaxConnWaitTimeout,
		"| MaxResponseBodySize: ", client.MaxResponseBodySize,
		"| DialTimeout: ", dialTimeout,
	)

	//pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
