package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
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

	minBodySize = 2000
)

var (
	jobs   int
	ScanWg sync.WaitGroup

	paths = []string{"/", "/wordpress", "/wp", "/blog", "/new", "/old", "/newsite", "/test", "/main", "/cms", "/dev", "/backup"}
)

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func main() {
	fmt.Println(banner)
	fmt.Println("jobs:", jobs)

	// Read domain list from specefied file
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	bDomains, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	domains := strings.Split(string(bDomains), "\n")
	//domainsLen := len(domains)

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

	// Init goroutines chan and run goroutines itself
	urlCh := make(chan string, jobs)
	resCh := make(chan Resp, jobs) // TODO read habr

	for i := 0; i < jobs; i++ {
		ScanWg.Add(1)
		go scan(urlCh, resCh)
	}

	// Response printer routine
	var RespWg sync.WaitGroup
	RespWg.Add(1)
	go func() {
		defer RespWg.Done()
		respCount := 0
		for resp := range resCh {
			respCount++
			switch resp.Status {
			case INSTALL:
				fmt.Printf(" Setup ===> %s\n", resp.Url)
				setupFile.WriteString(resp.Url + "\n")
			case SETUP:
				fmt.Printf(" Install ===> %s\n", resp.Url)
				installFile.WriteString(resp.Url + "\n")
			case FAIL:

			}
		}
	}()

	// Send http:// + domain to goroutine
	for _, domain := range domains {
		urlCh <- "http://" + domain
	}

	/*-------------------------------- 		END 	  --------------------------------*/

	close(urlCh)
	ScanWg.Wait()
	close(resCh)
	RespWg.Wait()

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
