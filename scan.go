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
   [+] Recoded By TiGER HeX [+] We Are TiGER HeX\n\n`
var usage = "USAGE:\tscan [--jobs jobsnum] FILE "

var jobs int
var wg sync.WaitGroup

/*-------------------------GOROUTINE'S ENTRYPOINT----------------------------*/
func scan(domains chan string, installWriter, setupWriter io.Writer) {
	for d := range domains {
		resp, err := http.Get(d)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			resp.Body.Close()
			continue
		}
		if strings.Contains(string(b), "WordPress &rsaquo; Installation") {
			fmt.Fprintf(installWriter, d+"\n")
			fmt.Println("\t[+] Install ===>", d)
		} else if strings.Contains(string(b), "WordPress &rsaquo; Setup Configuration File") {
			fmt.Fprintf(setupWriter, d+"\n")
			fmt.Println("\t[+] Setup ===>", d)
		} else {
			fmt.Println(" [+] Failed ===>", d)
		}
		resp.Body.Close()
	}
	wg.Done()
}

func main() {

	fmt.Println(banner)
	start := time.Now()

	/*---------------------------------FILEWORK---------------------------------------*/

	//Initializing domain file reader
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Printf(err.Error())
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
	domains := make(chan string)
	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go scan(domains, fi, fs)
	}
	/*--------------------------------------------------------------------------------*/

	/*-----------------------------SEND FILEDATA TO THEM------------------------------*/
	for {
		line, _, err := domainReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic("while reading domain list file")
		}
		domains <- string(line)
	}
	/*--------------------------------------------------------------------------------*/

	/*--------------------------------CLOSE CONNECTION--------------------------------*/
	close(domains)
	wg.Wait()
	fmt.Println("Time", time.Now().Sub(start))
	return
}

func init() {
	//setting jobs flag to parse with initial value of 48 goroutines
	flag.IntVar(&jobs, "jobs", 48, "number of goroutines to run")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println(usage)
		os.Exit(2)
	}
}
