package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
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
		MaxConnDuration:          time.Minute,
		MaxIdleConnDuration:      10 * time.Second,
		MaxResponseBodySize:      maxBodySize,
	}
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scan(urlCh, insCh, setCh, failCh, errCh chan string) {

	res := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()

	defer fasthttp.ReleaseResponse(res)
	defer fasthttp.ReleaseRequest(req)
	defer wg.Done()

	bodyBuff := make([]byte, maxBodySize)

	for url := range urlCh {
		req.SetRequestURI(url)

		err := client.DoRedirects(req, res, maxRedirects)
		if err != nil {
			failCh <- url
			errCh <- fmt.Sprintf("%s ===> %s\n", err.Error(), url)
			continue
		}

		//Save body from erase
		copy(bodyBuff, res.Body())
		//bodyBuff = res.Body()

		if res.StatusCode() != 200 {
			failCh <- fmt.Sprintf("%s [%d]", url, res.StatusCode())
			continue
		}

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

	/*--------------------------------------------------------------------------------*/
	/*--------------------------------START GOROUTINES--------------------------------*/

	start := time.Now()

	urlCh := make(chan string, jobs)
	insCh := make(chan string, jobs)
	setCh := make(chan string, jobs)
	failCh := make(chan string, jobs)
	errCh := make(chan string, jobs)

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scan(urlCh, insCh, setCh, failCh, errCh)
	}

	// Response writer goroutine
	go func() {
		for {
			select {
			case res := <-insCh:
				fmt.Printf("\t[+] Install ===> %s\n", res)
				installFile.WriteString(res + "\n")
			case res := <-setCh:
				fmt.Printf("\t[+] Setup ===> %s\n", res)
				setupFile.WriteString(res + "\n")
			case res := <-failCh:
				fmt.Printf(" [-] Fail ===> %s\n", res)
			case err := <-errCh:
				logger.Print(err)
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
	close(insCh)
	close(setCh)
	close(failCh)

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
