/*
File: loadgen.go
Author: Robinson Thompson

Description: Runs stress/performance tests on timeserver.go to verify functionality & performance

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "../seelog-master"
counter "../counter"
"flag"
"fmt"
//"html/template"
//"math/rand"
"net/http"
"os"
//"os/exec"
//"strconv"
//"strings"
"sync"
"time"
)

var requestRate *int
var requestBurst *int
var timeout *int
var programRuntime *int
var URL *string
var P2F int

var mutex = &sync.Mutex{}
var requestCodeMap = make(map[string]int)
var Counter *counter.Counter

/*
Setup for time.Ticker.  Runs infinitely and generates a burst of "requestBurst" concurrentrequests 
to the specified timeserver URL after a given time interval.
*/
func runTick() {
    ticker := time.NewTicker(time.Duration(*requestBurst * 1000000 / *requestRate) * time.Microsecond)
    for range ticker.C {
        //create "burst" go routines that each make 1 get request
        for i := 0; i < *requestBurst; i++ {
	    // generate 1 request
	    go generateRequest()
        }
    }
}

/*
Generates a request to the specified timeserver URL and records the response code as a 100 code, a
200 code, a 300 code, a 400 code, or a 500 code.  If the response code is a 408 timeout code, it
will be logged as an error code as well.
*/
func generateRequest() {
    Counter.Increment("Total")

    //credits to http://blog.golang.org/go-concurrency-patterns-timing-out-and
    getTimeout := make(chan bool, 1)

    client := http.Client {}
    timeResp := make(chan bool, 1)
    respCode := 0

    go func() {
	time.Sleep(time.Millisecond * time.Duration(*timeout))
	getTimeout <- true
    }()

    go func() {
	authResp, err := client.Get(*URL)
	respCode = authResp.StatusCode
	timeResp <- true
        if err != nil {
	    fmt.Println(err)
            if P2F == 1 {
	        defer Log.Flush()
    	        Log.Error("Error running client.Get from loadGen\r\n")
	    }
	    return
        }
    }()

    select {
	case <-getTimeout:
	    mutex.Lock()
	    //requestCodeMap["error"] = requestCodeMap["error"] + 1
	    Counter.Increment("Errors")
	    mutex.Unlock()
	    return
	case <-timeResp:
    }

    codeCentury := respCode / 100
    if codeCentury == 1 {
	mutex.Lock()
	//requestCodeMap["100"] = requestCodeMap["100"] + 1
	Counter.Increment("100s")
	mutex.Unlock()
    } else if codeCentury == 2 {
	mutex.Lock()
	//requestCodeMap["200"] = requestCodeMap["200"] + 1
	Counter.Increment("200s")
	mutex.Unlock()
    } else if codeCentury == 3 {
	mutex.Lock()
	//requestCodeMap["300"] = requestCodeMap["300"] + 1
	Counter.Increment("300s")
	mutex.Unlock()
    } else if codeCentury == 4 {
	mutex.Lock()
	//requestCodeMap["400"] = requestCodeMap["400"] + 1  
	Counter.Increment("400s") 
	mutex.Unlock()
    } else if codeCentury == 5 {
	mutex.Lock()
	//requestCodeMap["500"] = requestCodeMap["500"] + 1  
	Counter.Increment("500s") 
	mutex.Unlock()
    }

}

/*
Output the number of 100/200/300/400/500/error codes received to the console.
*/
func outputStatistics() {/*
    fmt.Println("Total Requests: " + strconv.Itoa(requestCodeMap["total"]))
    fmt.Println("100 Status Code Count: " + strconv.Itoa(requestCodeMap["100"]))
    fmt.Println("200 Status Code Count: " + strconv.Itoa(requestCodeMap["200"]))
    fmt.Println("300 Status Code Count: " + strconv.Itoa(requestCodeMap["300"]))
    fmt.Println("400 Status Code Count: " + strconv.Itoa(requestCodeMap["400"]))
    fmt.Println("500 Status Code Count: " + strconv.Itoa(requestCodeMap["500"]))
    fmt.Println("408 (Request Timeout) Status Code Count: " + strconv.Itoa(requestCodeMap["error"]))
    */

    fmt.Println(Counter.OutputKey("Total"))
    fmt.Println(Counter.OutputKey("100s"))
    fmt.Println(Counter.OutputKey("200s"))
    fmt.Println(Counter.OutputKey("300s"))
    fmt.Println(Counter.OutputKey("400s"))
    fmt.Println(Counter.OutputKey("500s"))
    fmt.Println(Counter.OutputKey("Errors"))
}

/*
Main
*/
func main() {
    fmt.Println("Beginning diagnostics")
    
    Counter = counter.New()
    Counter.AddKey("Total")
    Counter.AddKey("100s")
    Counter.AddKey("200s")
    Counter.AddKey("300s")
    Counter.AddKey("400s")
    Counter.AddKey("500s")
    Counter.AddKey("Errors")

    requestRate = flag.Int("rate", 1, "") 

    requestBurst = flag.Int("burst", 1, "")	 

    timeout = flag.Int("timeout-ms", 1000, "")

    programRuntime = flag.Int("runtime", 1, "")

    URL = flag.String("url", "http://localhost:8080/time", "")

    printToFile := flag.Bool("p2f", false, "")

    logPath := flag.String("log", "../../etc/seelog.xml", "")


    if *URL == "" {
	fmt.Println("No URL specified to test, now exiting program")
	os.Exit(1)
    }

    //Setup the seelog logger (cudos to http://sillycat.iteye.com/blog/2070140, https://github.com/cihub/seelog/blob/master/doc.go#L57)
    logger,loggerErr := Log.LoggerFromConfigAsFile(*logPath)
    if loggerErr != nil {
    	fmt.Println("Error creating logger from .xml log configuration file")
    } else {
	Log.ReplaceLogger(logger)
    }

    P2F = 0 // set to false
    flag.Parse()


    if *printToFile == true {
	P2F = 1 // set to true
    }

    go runTick()
    time.Sleep(time.Duration(*programRuntime) * time.Second)
    outputStatistics()
    Counter.GoDump()
    os.Exit(1)
}
