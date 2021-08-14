package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var banner = `
    _    _ ______       _____  _____  _____ 
   | |  | || ___ \     |_   _|/  ___|/  ___|
   | |  | || |_/ /______ | |  \ '--. \ '--. 
   | |/\| ||  __/|______|| |   '--. \ '--. \
   \  /\  /| |          _| |_ /\__/ //\__/ /
    \/  \/ \_|          \___/ \____/ \____/ 
                                            
   [+] WP Install and Setup Scanner
   [+] Recoded By TiGER HeX [+] We Are TiGER HeX`
var usage = "USAGE:\tscan [--jobs jobsnum] FILE "

var jobs int
var wg sync.WaitGroup

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scan(urls chan string, installWriter, setupWriter io.Writer) {

	tr := &http.Transport{
		IdleConnTimeout:    5 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	defer wg.Done()

	for url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			resp.Body.Close()
			continue
		}
		if strings.Contains(string(b), "WordPress &rsaquo; Installation") {
			fmt.Fprintln(installWriter, url)
			fmt.Println("\t[+] Install ===>", url)
		} else if strings.Contains(string(b), "WordPress &rsaquo; Setup Configuration File") {
			fmt.Fprintln(setupWriter, url)
			fmt.Println("\t[+] Setup ===>", url)
		} else {
			fmt.Println(" [+] Failed ===>", url)
		}
	}
}

func main() {

	fmt.Println(banner)
	start := time.Now()

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
	}
	defer fi.Close()

	//Creating setup.txt and initializing writer on it
	fs, err := os.OpenFile("setup.txt", os.O_CREATE+os.O_APPEND+os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer fs.Close()
	/*--------------------------------------------------------------------------------*/

	/*--------------------------------START GOROUTINES--------------------------------*/
	url_chan := make(chan string)
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scan(url_chan, fi, fs)
	}
	/*--------------------------------------------------------------------------------*/

	/*----------------------------SEND URLS+PATH TO THEM------------------------------*/

	paths := []string{"/", "/wordpress/", "/wp/", "/blog/", "/new/", "/old/", "/newsite/",
		"/test/", "/dev/", "/New/", "/Wp/", "/Wordpress/", "/Blog/", "/Newsite/", "/Dev/", "Test/", "Old/"}

	for {
		line, _, err := domainReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("while reading domain list file")
		}
		for _, path := range paths {
			url_chan <- "http://" + string(line) + path
		}
	}
	/*--------------------------------------------------------------------------------*/

	/*--------------------------------CLOSE CONNECTION--------------------------------*/
	close(url_chan)
	wg.Wait()
	fmt.Println("Time", time.Now().Sub(start))
	return
}

func init() {
	//setting jobs flag to parse with initial value of 48 goroutines
	flag.IntVar(&jobs, "jobs", 300, "number of goroutines to run")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
}
