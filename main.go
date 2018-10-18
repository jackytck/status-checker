package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// PageStatus stores the results of checking.
type PageStatus struct {
	Link     string
	Status   string
	Duration int64
}

func check(links []string, timeout int, verbose bool) []PageStatus {
	var wg sync.WaitGroup
	c := make(chan PageStatus)
	wg.Add(len(links))

	go func() {
		wg.Wait()
		close(c)
	}()

	for _, link := range links {
		go checkLink(link, timeout, c, &wg)
	}

	var ret []PageStatus
	for ps := range c {
		if verbose {
			fmt.Printf("%+v\n", ps)
		}
		ret = append(ret, ps)
	}

	return ret
}

func checkLink(link string, timeout int, c chan PageStatus, wg *sync.WaitGroup) {
	defer wg.Done()

	// setup http client with timeout
	client := http.Client{
		Timeout: time.Duration(time.Duration(timeout) * time.Second),
	}

	start := time.Now()
	_, err := client.Get(link)
	end := time.Now()
	elapsed := end.Sub(start)

	ps := PageStatus{
		Link:     link,
		Status:   "up",
		Duration: elapsed.Nanoseconds() / int64(time.Millisecond),
	}
	if err != nil {
		ps.Status = "down"
		if uerr, ok := err.(*url.Error); ok {
			if uerr.Timeout() {
				ps.Status = "timeout"
			}
		}
	}
	c <- ps
}

func main() {
	links := []string{
		"https://google.com",
		"https://facebook.com",
		"https://stackoverflow.com",
		"https://golang.com",
		"https://amazon.com",
	}
	timeout := 30 // in seconds

	results := check(links, timeout, true)
	fmt.Printf("%+v\n", results)
}
