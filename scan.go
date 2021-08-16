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
)

var (
	jobs int
	wg   sync.WaitGroup

	logger *log.Logger

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		MaxConnsPerHost:          20,
		MaxConnDuration:          10 * time.Second,
		MaxIdleConnDuration:      10 * time.Second,
		ReadTimeout:              10 * time.Second,
		WriteTimeout:             10 * time.Second,
	}
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scanner(urls chan string, installWriter, setupWriter io.Writer) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer wg.Done()
	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)

	req.SetConnectionClose()

	bodyBuff := make([]byte, maxBodySize)

	for url := range urls {
		req.SetRequestURI(url)
		err := client.DoRedirects(req, res, maxRedirects)

		if err != nil {
			fmt.Println(" [-] Bad Response ===>", url)
			logger.Printf("%v ===> %s\n", err.Error(), url)
			continue
		}

		code := res.StatusCode()
		bodySize := len(res.Body())

		if code != 200 || bodySize > maxBodySize {
			fmt.Println(" [-] Failed ===>", url, code)
			logger.Printf("Bad code(%d )or bodyLen(%d) ===> %s\n", code, bodySize, url)
			continue
		}

		//Save body from erase to body buffer
		copy(bodyBuff, res.Body())

		//Check this body contain pattern
		if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Installation") {
			fmt.Fprintln(installWriter, url)
			fmt.Println("\t[+] Install ===>", url, code, bodySize)
		} else if strings.Contains(string(bodyBuff), "WordPress &rsaquo; Setup Configuration File") {
			fmt.Fprintln(setupWriter, url)
			fmt.Println("\t[+] Setup ===>", url, code, bodySize)
		} else {
			if strings.Contains(url, "myfallretreat.conwaybcm.com/newsite") {
				fmt.Println("DEBUG ", bodyBuff)
			}
			fmt.Println(" [-] Failed ===>", url, code)
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

	url_chan := make(chan string, jobs)
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scanner(url_chan, fi, fs)
	}
	/*--------------------------------------------------------------------------------*/
	/*----------------------------READ & SEND URL+PATH TO THEM------------------------*/

	paths := []string{"/", "wordpress/", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "backup/"}

	for {
		line, _, err := domainReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("while reading domain list file")
			logger.Println(err)
			return
		}

		for _, path := range paths {
			url_chan <- "http://" + string(line) + path
		}
	}
	/*--------------------------------------------------------------------------------*/
	/*-------------------------------- 		END 	  --------------------------------*/
	close(url_chan)
	wg.Wait()
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
