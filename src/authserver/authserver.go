/*
File: authserver.go
Author: Robinson Thompson

Description: Authentication server for timeserver to get/set user cookies

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "../seelog-master"
"encoding/json"
"io/ioutil"
"flag"
"fmt"
"net/http"
"os"
"strconv"
//"strings"
"sync"
"time"
)
var authport *int
var hostname *string
var backupTime *int
var printToFile int
var mutex = &sync.Mutex{}
var cookieMap = make(map[string]http.Cookie)

/*
Handler for cookie "get" requests
Attmpts to lookup user name based on provided cookie UUID.  
Returns user name or empty with response code 200 on valid request.  
Returns empty string with response code 400 on malformed request
*/
func getRedirectHandler (w http.ResponseWriter, r *http.Request) {
    responseCode := 200

    r.ParseForm()
    cookieName := ""
    cookieUUID := r.FormValue("cookie")
    if cookieUUID == "" { 
	responseCode = 400 // set response code to 400, malformed request
    } else {
	responseCode = 200 // set response code to 200, request processed
    }
     
    //Attempt to retrieve user name from cookie map based on UUID
    foundCookie := false

    mutex.Lock()
    cookieLookup := cookieMap[cookieUUID]
    mutex.Unlock()

    if cookieLookup.Name != "" {
	foundCookie = true
	cookieName = cookieLookup.Value
    }

    if !foundCookie {
	responseCode = 400 // set response code to 400, malformed request
    }
     
    w.WriteHeader(responseCode)
    w.Write([]byte(cookieName))
    // timeserver will need to use r.ParseForm() and http.get(URL (i.e. authhost:9090/get) to retrieve data
}

/*
Handler for cookie "set" requests
Attempts to add a new provided cookie to internal cookie map
Returns response code 200 on processed request
Returns response code 400 on malformed request
*/
func setRedirectHandler (w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    cookieTemp := r.FormValue("cookie")
    cookieName := r.FormValue("name") // UUID (used as cookieMap's cookie name)

    if cookieTemp == "" || cookieName == "" {
	w.WriteHeader(400) // set response code to 400, request malformed
	return
    } else {
        w.WriteHeader(200) // set response code to 200, request processed
    }

    // attempt to add cookie to internal cookie map
    var newCookie http.Cookie
    err1 := json.Unmarshal([]byte(cookieTemp), &newCookie)
    if err1 != nil {
        fmt.Println("Error unmarshalling new cookie")

        if printToFile == 1 {
	    defer Log.Flush()
	    Log.Error("Error unmarshalling new cookie")
	    return
	}
    }
    
    mutex.Lock()
    cookieMap[cookieName] = newCookie
    mutex.Unlock()
}


/*
Handler for cookie "clear" requests
Attempts to clear a user cookie from cookieMap
Returns response code 200 on processed request
Returns response code 400 on malformed request
*/
func clearRedirectHandler (w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    cookieName := r.FormValue("cookie") // UUID (used as cookieMap's cookie name)
    if cookieName == ""{
	w.WriteHeader(400) // set response code to 400, request malformed
	return
    } else {
        delete(cookieMap, cookieName) // delete cookie from map (if exists)
        w.WriteHeader(200) // set response code to 200, request processed
    }
}


/*

*/
func errorhandler (w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(404) // Set response code to 404 not found
}

/*
Updates dumpfile.txt 
*/
func Updatedumpfile() {
    stallDuration := time.Duration(*backupTime)*time.Second
    for {
	time.Sleep(stallDuration)
        mutex.Lock()
        encodedMap,encodeErr := json.Marshal(cookieMap)
        mutex.Unlock()

        if encodeErr != nil {
            fmt.Println("Error JSON encoding cookie map")

            if printToFile == 1 {
	        defer Log.Flush()
	        Log.Error("Error JSON encoding cookie map/r/n")
	        return
	    }
    	}

        oldDump,err := ioutil.ReadFile("dumpfile.txt")
        if err != nil { //Assume that dumpfile.txt hasn't been made yet
            ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
	    readCopy,err2 := ioutil.ReadFile("dumpfile.txt")
	    if err2 != nil {
	       fmt.Println("Error reading dumpfile")
                if printToFile == 1 {
		    defer Log.Flush()
		    Log.Error("Error reading dumpfile/r/n")
		    return
	        }
            }
	    mutex.Lock()
	    err3 := json.Unmarshal(readCopy, &cookieMap)
	    mutex.Unlock()
	    if err3 != nil {
	        fmt.Println("Error unmarshaling")
	        if printToFile == 1 {
	    	    defer Log.Flush()
		    Log.Error("Error unmarshaling/r/n")
	        }
            }
        }

        ioutil.WriteFile("dumpfile.bak", oldDump, 0644)
        os.Remove("dumpfile.txt")
        ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
        readCopy,err5 := ioutil.ReadFile("dumpfile.txt")
        if err5 != nil {
	    fmt.Println("Error reading dumpfile")
            if printToFile == 1 {
	        defer Log.Flush()
	        Log.Error("Error reading dumpfile/r/n")
	        return
	    }
        }

        mutex.Lock()
        err6 := json.Unmarshal(readCopy, &cookieMap)
        mutex.Unlock()
        if err6 != nil {
	    fmt.Println("Error unmarshaling")
	    if printToFile == 1 {
                defer Log.Flush()
	        Log.Error("Error unmarshaling/r/n")
	    }
	    return
        }
        os.Remove("dumpfile.bak")
    }
}

/*
Main
*/
func main() {
    fmt.Println("Now starting authentication server")

    p2f := flag.Bool("p2f", false, "") //flag to output to file

    logPath := flag.String("log", "../../etc/seelog.xml", "")

    authport = flag.Int("authport", 9090, "")

    hostname = flag.String("hostname", "localhost:", "")

    loadDumpFile := flag.Bool("dumpfile", false, "")

    backupTime = flag.Int("checkpoint-interval", 0, "")

    flag.Parse()

    printToFile = 0 // set to false
    if *p2f == true {
	printToFile = 1 // set to true
    }


    if *loadDumpFile == true {
	dump,err := ioutil.ReadFile("dumpfile.txt")
	if err != nil {
	    fmt.Println("Error reading dumpfile")
            if printToFile == 1 {
		defer Log.Flush()
		Log.Error("Error reading dumpfile")
	    }
        } else {
	    mutex.Lock()
	    err2 := json.Unmarshal(dump, &cookieMap)
	    mutex.Unlock()
	    if err2 != nil {
	        fmt.Println("Error unmarshaling")
	        if printToFile == 1 {
		    defer Log.Flush()
		    Log.Error("Error unmarshaling")
	        }
            }
        }
    }

    if *backupTime > 0 {
	go Updatedumpfile()
    }

    //Setup the seelog logger (cudos to http://sillycat.iteye.com/blog/2070140, https://github.com/cihub/seelog/blob/master/doc.go#L57)
    logger,loggerErr := Log.LoggerFromConfigAsFile(*logPath)
    if loggerErr != nil {
    	fmt.Println("Error creating logger from .xml log configuration file")
    } else {
	Log.ReplaceLogger(logger)
    }

    http.HandleFunc("/", errorhandler)
    http.HandleFunc("/get", getRedirectHandler)
    http.HandleFunc("/set", setRedirectHandler)
    http.HandleFunc("/clear", clearRedirectHandler)

    error := http.ListenAndServe(*hostname + strconv.Itoa(*authport), nil)
    if error != nil {				// If the specified port is already in use, 
	fmt.Println("Port already in use")	// output a error message and exit with a 
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Error("Port already in use\r\n")
        }
	os.Exit(1)				// non-zero error code
    }
}
