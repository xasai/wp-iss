package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
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
												
	   [+] WP Install and Setup Scanner
	   [+] Recoded By TiGER HeX [+] We Are TiGER HeX
	   `
	usage = "USAGE:\tscan [--jobs jobs_num] FILE "

	maxRedirects = 20

	maxBodySize = 300000
	minBodySize = 2000
)

var (
	//Flags
	jobs        int   = 1000 //default value of threads ! change this
	dialTimeout int64 = 3    //seconds

	reqTotal uint64
	ScanWg   sync.WaitGroup
	logger   *log.Logger
	errlog   *log.Logger
	client   *fasthttp.Client
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
	rdata, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	urls := strings.Split(strings.TrimRight(strings.Replace(string(rdata), "\r\n", "\n", -1), "\n"), "\n")

	// Create file to write install result
	installFile, err := os.OpenFile("install.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer installFile.Close()

	// Create file to write setup result
	setupFile, err := os.OpenFile("setup.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer setupFile.Close()

	start := time.Now()

	//main will send urls from domain list for scanners routines
	urlCh := make(chan string)

	//printer routine will receive result from all scanners routines
	resCh := make(chan *Resp, 10)

	//launch scanners
	ScanWg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go scan(urlCh, resCh)
	}

	//launch printer routine
	var RespWg sync.WaitGroup
	go func() {
		RespWg.Add(1)
		defer RespWg.Done()

		bar := progressbar.NewOptions(
			len(urls),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionThrottle(time.Millisecond*60),
		)
		defer bar.Close()
		for resp := range resCh {
			switch resp.Status {
			case INSTALL:
				bar.Clear()
				fmt.Printf(" [+] Install ===> %s\n", resp.Url)
				logger.Printf(" Install %s\n", resp.Url)
				bar.Add(1)
				installFile.WriteString(resp.Url + "\n")
			case SETUP:
				bar.Clear()
				fmt.Printf(" [+] Setup ===> %s\n", resp.Url)
				logger.Printf(" Setup %s\n", resp.Url)
				bar.Add(1)
				setupFile.WriteString(resp.Url + "\n")
			case FAIL:
				bar.Add(1)
			}
		}
	}()

	// Send http:// + url to scanners routines
	for _, url := range urls {
		urlCh <- "http://" + url
	}

	//Wait scanners end
	close(urlCh)
	ScanWg.Wait()
	//Wait printer end
	close(resCh)
	RespWg.Wait()

	fmt.Println("", time.Now().Sub(start), "Requests total:", reqTotal)
	logger.Println("", time.Now().Sub(start), "Requests total:", reqTotal,
		"")
}

func init() {

	var logOn bool

	flag.IntVar(&jobs, "j", jobs, "number of goroutines to run")
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
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*time.Duration(dialTimeout))
		},
	}

	if logOn {
		initLoggers()
	} else {
		logger = log.New(ioutil.Discard, "", 0)
		errlog = log.New(ioutil.Discard, "", 0)
	}
}

func initLoggers() {
	logf, err := os.OpenFile(flag.Arg(0)+"_bench", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	logger = log.New(logf, "", 0)

	logf, err = os.OpenFile("errlog", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	errlog = log.New(logf, "", 0)

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
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}