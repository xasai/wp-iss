package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

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
	minBodySize  = 2000
)

var (
	jobs int
	wg   sync.WaitGroup

	logger *log.Logger

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		MaxConnDuration:          10 * time.Second,
		MaxIdleConnDuration:      10 * time.Second,
		MaxResponseBodySize:      maxBodySize,
	}
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scanner(urlCh, resultCh, errCh chan string) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer wg.Done()

	bodyBuff := make([]byte, maxBodySize)

	for url := range urlCh {
		req.SetRequestURI(url)
		req.SetConnectionClose()

		err := client.DoRedirects(req, res, maxRedirects)
		if err != nil {
			resultCh <- fmt.Sprintf(" [-] Bad Response ===> %s\n", url)
			errCh <- fmt.Sprintf("%s ===> %s\n", err.Error(), url)
			continue
		}

		//Save body from erase
		copy(bodyBuff, res.Body())

		code := res.StatusCode()
		if code != 200 {
			resultCh <- fmt.Sprintf(" [-] Failed ===> %s %d\n", url, code)
			continue
		}

		if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Installation") {
			resultCh <- fmt.Sprintf("\t[+] Install ===> %s %d\n", url, code)
		} else if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Setup Configuration File") {
			resultCh <- fmt.Sprintf("\t[+] Setup ===> %s %d\n", url, code)
		} else {
			resultCh <- fmt.Sprintf(" [-] Failed ===> %s %d\n", url, code)
		}
	}
}

func main() {

	fmt.Println(banner)
	fmt.Println("jobs:", jobs)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	/*---------------------------------FILEWORK---------------------------------------*/

	//Initializing domain file reader
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	defer f.Close()
	domainReader := bufio.NewReader(f)

	//Creating install.txt and initializing writer on it

	fi, err := os.OpenFile("install.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer fi.Close()

	//Creating setup.txt and initializing writer on it

	fs, err := os.OpenFile("setup.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer fs.Close()

	/*--------------------------------------------------------------------------------*/
	/*--------------------------------START GOROUTINES--------------------------------*/

	start := time.Now()

	urlCh := make(chan string, jobs)
	resCh := make(chan string, jobs)
	errCh := make(chan string, jobs)

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scanner(urlCh, resCh, errCh)
	}

	// Response writer goroutine
	//wg.Add(1)
	go func() {
		//defer wg.Done()
		for {
			select {
			case err := <-errCh:
				logger.Print(err)
			case res := <-resCh:
				fmt.Print(res)
			default:
				break
			}
		}
	}()

	/*--------------------------------------------------------------------------------*/
	/*----------------------------READ & SEND URL+PATH TO THEM------------------------*/

	paths := []string{"/", "/wordpress", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "/backup"}

	for {
		line, _, err := domainReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error occured while reading domain list file")
			logger.Println(err)
			return
		}
		for _, path := range paths {
			urlCh <- "http://" + string(line) + path
		}
	}

	/*--------------------------------------------------------------------------------*/
	/*-------------------------------- 		END 	  --------------------------------*/

	close(urlCh)
	wg.Wait()
	close(errCh)
	close(resCh)
	fmt.Println("Time", time.Now().Sub(start))
}

func init() {
	//setting jobs flag to parse with initial value of 48 goroutines
	flag.IntVar(&jobs, "jobs", 300, "number of goroutines to run")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
	l, err := os.OpenFile("error.log", os.O_WRONLY+os.O_CREATE+os.O_TRUNC, 0660)
	if err != nil {
		logger = log.Default()
	} else {
		logger = log.New(l, "", 0)
	}
}
