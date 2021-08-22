package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"log"

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
	usage = "USAGE:\tscan [--jobs jobsnum] FILE "

	maxRedirects = 20
	maxBodySize  = 15000

	minBodySize = 2000
)

var (
	jobs     int
	reqTotal uint64
	ScanWg   sync.WaitGroup

	logger *log.Logger
	client *fasthttp.Client
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func main() {
	fmt.Println(banner)
	fmt.Println(" jobs:", jobs)

	//PPROF
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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
	urls := strings.Split(string(rdata), "\n")

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
	resCh := make(chan *Resp, jobs)

	//launch scanners
	for i := 0; i < jobs; i++ {
		ScanWg.Add(1)
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
			progressbar.OptionThrottle(time.Second/30),
		)

		for resp := range resCh {
			switch resp.Status {
			case INSTALL:
				bar.Clear()
				fmt.Printf(" [+] Install ===> %s\n", resp.Url)
				bar.Add(1)
				setupFile.WriteString(resp.Url + "\n")
			case SETUP:
				bar.Clear()
				fmt.Printf(" [+] Setup ===> %s\n", resp.Url)
				bar.Add(1)
				installFile.WriteString(resp.Url + "\n")
			case FAIL:
				bar.Add(1)
			}
		}
		bar.Close()
	}()

	// Send http:// + url to scanners routines
	for _, url := range urls {
		urlCh <- "http://" + url
	}

	//Wait for scanners end
	close(urlCh)
	ScanWg.Wait()

	//Wait for printer end
	close(resCh)
	RespWg.Wait()

	fmt.Println("", time.Now().Sub(start), "Requests total:", reqTotal)
}

func init() {
	flag.IntVar(&jobs, "jobs", 100, "number of goroutines to run")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
	logf, _ := os.Create(".miss")
	logger = log.New(logf, "", 0)

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		MaxConnDuration:          7 * time.Second, //check this out FIXME
		MaxIdleConnDuration:      7 * time.Second,
		ReadTimeout:              7 * time.Second,
		WriteTimeout:             7 * time.Second,
		MaxResponseBodySize:      maxBodySize,
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, 7*time.Second)
		},
	}
}
