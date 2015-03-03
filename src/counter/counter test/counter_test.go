/*
File: counter_test.go
Author: Robinson Thompson

Description: Testing for counter that allows thread-safe incrementing of ints

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package counter_test

import (
"encoding/json"
"io/ioutil"
"fmt"
"os"
"strconv"
"sync"
//"sync/atomic"
"testing"
)

var key = "100s"



type Counter struct {
    mutex sync.Mutex
    counterMap map[string]int
}


/*
TEST VERSION
Creates a new Counter and returns it.
*/
func TestNew(tester *testing.T){

}


/*
TEST VERSION
Adds a new key to the Counter's map.  WILL CLEAR KEY CONTENTS IF KET ALREADY EXISTS
*/
func TestAddKey(tester *testing.T) {
    c := &Counter {
        mutex:  sync.Mutex{},
        counterMap: make(map[string]int),
    }

    c.mutex.Lock()
    c.counterMap[key] = 0
    c.mutex.Unlock()
}

/*
TEST VERSION
Increments the key indicated by "key" in c Counter's counterMap by 1.  Generates the 
key in counterMap if it doesn't already exist.
*/
func TestIncrement(tester *testing.T) {
    c := &Counter {
        mutex:  sync.Mutex{},
        counterMap: make(map[string]int),
    }

    _, wasFound := c.counterMap[key]

    if wasFound {
	c.mutex.Lock()
	c.counterMap[key] += 1
	c.mutex.Unlock()
    }else {
	c.mutex.Lock()
	c.counterMap[key] = 1
	c.mutex.Unlock()
    }
}

/*
TEST VERSION
Output key's contents in counterMap.
*/
func TestOutputKey(tester *testing.T){
}

/*
TEST VERSION
Output everything in counterMap.

NOTE: May not always be in order
*/
func TestOutputMap(tester *testing.T) {
    c := &Counter {
        mutex:  sync.Mutex{},
        counterMap: make(map[string]int),
    }

    output := ""

    for currCount := range c.counterMap {
	output += currCount + ": " + strconv.Itoa(c.counterMap[currCount]) + "\n"
    }
}

/*
TEST VERSION
Updates dumpfile.txt 
*/
func TestGoDump(tester *testing.T) {
    c := &Counter {
        mutex:  sync.Mutex{},
        counterMap: make(map[string]int),
    }

        c.mutex.Lock()
        encodedMap,encodeErr := json.Marshal(c.counterMap)
        c.mutex.Unlock()

        if encodeErr != nil {
            fmt.Println("Error JSON encoding counter map")
	    tester.Error("Error JSON encoding counter map")
    	}

        oldDump,err := ioutil.ReadFile("dumpfile.txt")
        if err != nil { //Assume that dumpfile.txt hasn't been made yet
            ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
	    readCopy,err2 := ioutil.ReadFile("dumpfile.txt")
	    if err2 != nil {
	       fmt.Println("Error reading dumpfile")
	       tester.Error("Error reading dumpfile")
            }
	    c.mutex.Lock()
	    err3 := json.Unmarshal(readCopy, &c.counterMap)
	    c.mutex.Unlock()
	    if err3 != nil {
	        fmt.Println("Error unmarshaling")
		tester.Error("Error unmarshaling")
            }
        }

        ioutil.WriteFile("dumpfile.bak", oldDump, 0644)
        os.Remove("dumpfile.txt")
        ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
        readCopy,err5 := ioutil.ReadFile("dumpfile.txt")
        if err5 != nil {
	    fmt.Println("Error reading dumpfile")
	    tester.Error("Error reading dumpfile")
        }

        c.mutex.Lock()
        err6 := json.Unmarshal(readCopy, &c.counterMap)
        c.mutex.Unlock()
        if err6 != nil {
	    fmt.Println("Error unmarshaling")
	    tester.Error("Error unmarshaling")
	    return
        }
        os.Remove("dumpfile.bak")
}