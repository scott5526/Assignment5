/*
File: counter.go
Author: Robinson Thompson

Description: Counter that allows thread-safe incrementing of ints

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package counter

import (
"encoding/json"
"io/ioutil"
"fmt"
"os"
"strconv"
"sync"
//"sync/atomic"
)

type Counter struct {
    mutex sync.Mutex
    counterMap map[string]int
}

/*
Creates a new Counter and returns it.
*/
func New() *Counter {
    newMap := make(map[string]int)
    ret := &Counter {
        mutex:  sync.Mutex{},
        counterMap: newMap,
    }

    return ret
}


/*
Adds a new key to the Counter's map.  WILL CLEAR KEY CONTENTS IF KET ALREADY EXISTS
*/
func (c *Counter) AddKey(key string) {
    c.mutex.Lock()
    c.counterMap[key] = 0
    c.mutex.Unlock()
}

/*
Increments the key indicated by "key" in c Counter's counterMap by 1.  Generates the 
key in counterMap if it doesn't already exist.
*/
func (c *Counter) Increment(key string) {
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
Output key's contents in counterMap.
*/
func (c *Counter) OutputKey(key string) string{
    return key + ": " + strconv.Itoa(c.counterMap[key])
}



/*
Output everything in counterMap.

NOTE: May not always be in order
*/
func (c *Counter) OutputMap() string{
    output := ""

    for currCount := range c.counterMap {
	output += currCount + ": " + strconv.Itoa(c.counterMap[currCount]) + "\n"
    }

    return output
}

/*
Updates dumpfile.txt 
*/
func (c *Counter) GoDump() {
        c.mutex.Lock()
        encodedMap,encodeErr := json.Marshal(c.counterMap)
        c.mutex.Unlock()

        if encodeErr != nil {
            fmt.Println("Error JSON encoding counter map")
    	}

        oldDump,err := ioutil.ReadFile("dumpfile.txt")
        if err != nil { //Assume that dumpfile.txt hasn't been made yet
            ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
	    readCopy,err2 := ioutil.ReadFile("dumpfile.txt")
	    if err2 != nil {
	       fmt.Println("Error reading dumpfile")
            }
	    c.mutex.Lock()
	    err3 := json.Unmarshal(readCopy, &c.counterMap)
	    c.mutex.Unlock()
	    if err3 != nil {
	        fmt.Println("Error unmarshaling")
            }
        }

        ioutil.WriteFile("dumpfile.bak", oldDump, 0644)
        os.Remove("dumpfile.txt")
        ioutil.WriteFile("dumpfile.txt", encodedMap, 0644)
        readCopy,err5 := ioutil.ReadFile("dumpfile.txt")
        if err5 != nil {
	    fmt.Println("Error reading dumpfile")
        }

        c.mutex.Lock()
        err6 := json.Unmarshal(readCopy, &c.counterMap)
        c.mutex.Unlock()
        if err6 != nil {
	    fmt.Println("Error unmarshaling")
	    return
        }
        os.Remove("dumpfile.bak")
}