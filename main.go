package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// PageStatus stores the results of checking.
type PageStatus struct {
	Link     string
	Status   string
	Duration int64
}

func check(links []string) {
	var wg sync.WaitGroup
	c := make(chan PageStatus)
	wg.Add(len(links))

	go func() {
		wg.Wait()
		close(c)
	}()

	for _, link := range links {
		go checkLink(link, c, &wg)
	}

	for ps := range c {
		fmt.Printf("%+v\n", ps)
	}
}

func checkLink(link string, c chan PageStatus, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	_, err := http.Get(link)
	end := time.Now()
	elapsed := end.Sub(start)

	ps := PageStatus{
		Link:     link,
		Status:   "up",
		Duration: elapsed.Nanoseconds() / int64(time.Millisecond),
	}
	if err != nil {
		ps.Status = "down"
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
	check(links)
}
