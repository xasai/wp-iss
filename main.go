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
	domainsLen := len(domains)

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

	// Setup goroutines chan and run goroutines
	urlCh := make(chan string, jobs)
	insCh := make(chan string, jobs)
	setCh := make(chan string, jobs)
	failCh := make(chan string, jobs)

	for i := 0; i < jobs; i++ {
		ScanWg.Add(1)
		go scan(urlCh, insCh, setCh, failCh)
	}

	// Response printer routine

	completed := 0
	var printer sync.WaitGroup
	printer.Add(1)
	go func() {
		defer printer.Done()
		for completed < domainsLen {
			select {
			case res := <-insCh:
				completed++
				fmt.Printf("![%d/%d] Install ===> %s\n", completed, domainsLen, res)
				installFile.WriteString(res + "\n")
			case res := <-setCh:
				completed++
				fmt.Printf("![%d/%d] Setup ===> %s\n", completed, domainsLen, res)
				setupFile.WriteString(res + "\n")
			case res := <-failCh:
				completed++
				fmt.Printf(" [%d/%d] Fail ===> %s\n", completed, domainsLen, res)
			}
		}
	}()

	// Send domain to goroutine
	for _, domain := range domains {
		urlCh <- "http://" + domain
	}

	/*--------------------------------------------------------------------------------*/
	/*-------------------------------- 		END 	  --------------------------------*/

	close(urlCh)
	ScanWg.Wait()
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
