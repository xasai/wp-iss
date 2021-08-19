package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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

	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true, // Don't send: User-Agent: fasthttp
		MaxConnDuration:          time.Minute,
		MaxIdleConnDuration:      10 * time.Second,
		MaxResponseBodySize:      maxBodySize,
	}

	paths = []string{"/", "/wordpress", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "/backup"}
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scan(urlCh, insCh, setCh, failCh chan string) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer wg.Done()

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

func main() {

	fmt.Println(banner)
	fmt.Println("jobs:", jobs)

	/*---------------------------------FILEWORK---------------------------------------*/

	//Initializing domain file reader
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	defer f.Close()
	domainReader := bufio.NewReader(f)

	installFile, err := os.OpenFile("install.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer installFile.Close()

	setupFile, err := os.OpenFile("setup.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer setupFile.Close()

	// Read domain file
	var lines []string
	var linesCount int

	fmt.Println("Reading file", flag.Arg(0), "...")
	for {
		line, _, err := domainReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error occured while reading domain list file")
			return
		}
		linesCount++
		lines = append(lines, string(line))
	}

	//Count urls to scan
	urlsCount := linesCount * (len(paths))

	/*--------------------------------------------------------------------------------*/
	/*--------------------------------START GOROUTINES--------------------------------*/

	start := time.Now()

	urlCh := make(chan string, jobs)
	insCh := make(chan string, jobs)
	setCh := make(chan string, jobs)
	failCh := make(chan string, jobs)

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scan(urlCh, insCh, setCh, failCh)
	}

	// Response printer routine

	completed := 0
	var printer sync.WaitGroup
	printer.Add(1)
	go func() {
		defer printer.Done()
		for completed < urlsCount {
			select {
			case res := <-insCh:
				completed++
				fmt.Printf("![%d/%d] Install ===> %s\n", completed, urlsCount, res)
				installFile.WriteString(res + "\n")
			case res := <-setCh:
				completed++
				fmt.Printf("![%d/%d] Setup ===> %s\n", completed, urlsCount, res)
				setupFile.WriteString(res + "\n")
			case res := <-failCh:
				completed++
				fmt.Printf(" [%d/%d] Fail ===> %s\n", completed, urlsCount, res)
			}
		}
	}()

	/*--------------------------------------------------------------------------------*/
	/*----------------------------SEND SCHEMA+URL+PATH -------------------------------*/

	for _, line := range lines {
		for _, path := range paths {
			urlCh <- "http://" + line + path
		}
	}

	/*--------------------------------------------------------------------------------*/
	/*-------------------------------- 		END 	  --------------------------------*/

	close(urlCh)
	wg.Wait()
	printer.Wait()
	close(insCh)
	close(setCh)
	close(failCh)

	fmt.Println("Time", time.Now().Sub(start))
}

func init() {
	//setting jobs flag to parse with initial value of 48 goroutines
	flag.IntVar(&jobs, "jobs", 100, "number of goroutines to run")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
}
