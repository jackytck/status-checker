package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// PageStatus stores the results of checking.
type PageStatus struct {
	Link     string `json:"link"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
}

func gatewayHandler(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ps, err := lambdaHandler()
	if err != nil {
		return serverError(err)
	}
	js, err := json.Marshal(ps)
	if err != nil {
		return serverError(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	fmt.Println(err)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

func lambdaHandler() ([]PageStatus, error) {
	links := []string{
		"https://google.com",
		"https://facebook.com",
		"https://stackoverflow.com",
		"https://golang.com",
		"https://amazon.com",
	}
	timeout := 30 // in seconds
	return check(links, timeout, true), nil
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
	lambda.Start(gatewayHandler)
}
