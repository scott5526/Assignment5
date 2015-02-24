/*
File: cookieManagement.go
Author: Robinson Thompson

Description:  Manages cookies for timeserver.go

Copyright:  All code was written originally by Robinson Thompson with assistance from various
	    free online resourses.  To view these resources, check out the README
*/
package main

import (
Log "./seelog-master"
"encoding/json"
"fmt"
"html/template"
"io/ioutil"
"net/http"
"net/url"
"strconv"
"time"
)

//Attempt to find a cookie on the user's browser and greet them using their name stored on it
func greetingCheck (w http.ResponseWriter, r *http.Request) {
    mutex.Lock()
    redirect = true
    mutex.Unlock()
    for _, currCookie := range r.Cookies() { // check all potential cookies stored by the user for a matching cookie
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Value
	    name, wasSuccess := authGet(w, r, currCookieVal)

            if wasSuccess {
    		fmt.Fprintf(w, "Greetings, " + name)
		mutex.Lock()
		redirect = false
		mutex.Unlock()
	    }
	}
    }
}

//Ensuring the user does not already have a browser cookie matching a cookie in the local cookie map, if they do
//redirect the user to the greetings page
func loginCheck (w http.ResponseWriter, r *http.Request) {
    for _, currCookie := range r.Cookies() {  //Run through the range of applicable cookies on the user's browser
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Value
	    _, wasSuccess := authGet(w, r, currCookieVal)

       	    if wasSuccess {
		path := *templatesPath + "greetingRedirect.html"
    		newTemplate2,err := template.New("redirect").ParseFiles(path)  
    		if err != nil {
		    fmt.Println("Error running get name redirect template")
        	    if printToFile == 1 {
			defer Log.Flush()
    			Log.Error("Error running greeting redirect template\r\n")
		    }
		    return
    		}  

    		newTemplate2.ExecuteTemplate(w,"greetingRedirectTemplate",portInfoStuff)
	    } else {
		currCookie.MaxAge = -1 //Invalidate user's browser cookie, it wasn't found in the authserver cookie map
		clearMapCookie(w,r)
	    }
    	}
     }
}

//Clears a user's cookie from the cookie map and clear (1) copy of the cookie from the user's browser (cannot delete
//copies the user may have created themselves and stored elsewhere)
func clearMapCookie (w http.ResponseWriter, r *http.Request) {
   mutex.Lock()
   redirect = false // set to true if user cookie is found (they are actually logged in)
   mutex.Unlock()

   for _, currCookie := range r.Cookies() {  //Run through the range of applicable cookies on the user's browser
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Value
	    _, wasSuccess := authGet(w, r, currCookieVal)

            if wasSuccess {
	        mutex.Lock()
	        redirect = true // user was actually logged in
	        mutex.Unlock()

	        authClear(w,r,currCookieVal)
	        currCookie.MaxAge = -1 //Set the user's cookie's MaxAge to an invalid number to expire it
	    } else {
		currCookie.MaxAge = -1 //Invalidate user's browser cookie, it wasn't found in the authserver cookie map
	    }
    	}
    }
}

//Creates a cookie for the user and adds it to their browser and the internal cookie map
func cookieSetup (w http.ResponseWriter, r *http.Request, newUUID string, expDate time.Time) {
    aNewCookie := http.Cookie{Name: "localhost", Value: newUUID, Expires: expDate, HttpOnly: true, MaxAge: 100000, Path: "/"}
    http.SetCookie(w, &aNewCookie)

    path := *templatesPath + "login.html"
    newTemplate,err := template.ParseFiles(path)   
    if err != nil {
	fmt.Println("Error running login template")
        if printToFile == 1 {
		defer Log.Flush()
    		Log.Error("Error running login template\r\n")
	}
	return;
    } 
    newTemplate.Execute(w,"loginTemplate")

    r.ParseForm()
    name := r.PostFormValue("name")
    submit := r.PostFormValue("submit")

    if submit == "Submit"{ // check if the user hit the "submit" button
    	if name == "" {
	    path = *templatesPath + "/badLogin.html"
    	    newTemplate,_ := template.New("outputUpdate").ParseFiles(path)   
    	    newTemplate.ExecuteTemplate(w,"badLoginTemplate",nil)
    	} else {
	    //generate cookie map's cookie
	    mapCookie := http.Cookie{
	        Name: newUUID, 
	        Value: name, 
	        Path: "/", 
		Domain: "localhost", 
		Expires: expDate,
 		HttpOnly: true, 
		MaxAge: 100000,
	    }
	    authSet(w, r, mapCookie, newUUID)

	    fmt.Println("localhost:" + strconv.Itoa(*portNO) + "/login?name=" + name)
    	    if  printToFile == 1 { //Check if the p2f flag is set
		defer Log.Flush()
    		Log.Info("localhost:" + strconv.Itoa(*portNO) + "/login?name=" + name + "\r\n")
    	    }

	    //Redirect to greetings (home) page
	    path = *templatesPath + "greetingRedirect.html"
    	    newTemp,err := template.New("redirect").ParseFiles(path)   
    	    if err != nil {
		fmt.Println("Error running greeting redirect template")
        	if printToFile == 1 {
		    defer Log.Flush()
    		    Log.Error("Error running greeting redirect template\r\n")
		}
		return;
    	    } 

    	    newTemp.ExecuteTemplate(w,"greetingRedirectTemplate",portInfoStuff)
    	}
    } else {
	//Form not submitted or incompatible browser (I'm looking at you Internet Explorer)
    }
}

//Retrieves the user name from their cookie on their browser (if one exists)
func getUserName (w http.ResponseWriter, r *http.Request) string {
    for _, currCookie := range r.Cookies() { //Lookup the user name by cross matching the user cookie's value against the local cookie maps's cookie names
    	if (currCookie.Name != "") {
	    currCookieVal := currCookie.Value
	    name, wasSuccess := authGet(w, r, currCookieVal)

            if wasSuccess {
    	        return ", " + name
	    } else {
		currCookie.MaxAge = -1 //Invalidate user's browser cookie, it wasn't found in the authserver cookie map
		clearMapCookie(w,r)
	    }
    	}
    }
    return ""
}

/*
Attempts to retrieve a user name from the cookie map stored in authserver.  Returns a string that
may or may not be empty + true on a successful lookup and an empty string + false on a failed lookup
*/
func authGet (w http.ResponseWriter, r *http.Request, userUUID string) (string, bool){
    returnName := ""
    getTimeout := time.Duration((time.Second / 1000) * time.Duration(*timeout))
    client := http.Client {
	Timeout: getTimeout,
    }

    getURL := "http://" + *authhost + strconv.Itoa(*authport) + "/get?cookie=" + userUUID
    authResp, err := client.Get(getURL)

    if err != nil {
	fmt.Println("Error running http.Get to retrieve name from authserver")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Error("Error running http.Get to retrieve name from authserver\r\n")
	}
	return returnName, false
    } else if authResp.StatusCode != 200 {
	//fmt.Println("authserver get returned non-200 status code")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Warn("authserver get returned non-200 status code\r\n")
	}
	return returnName, false
    }

    authBody, err2 := ioutil.ReadAll(authResp.Body)
    defer authResp.Body.Close()
    if err2 != nil {
	fmt.Println("Error reading authserver response body using ioutil")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Error("Error reading authserver response body using ioutil\r\n")
	}
    }
    
    returnName = string(authBody[:])
    return returnName, true
}

//Adds a cookie to the cookie map
func authSet (w http.ResponseWriter, r *http.Request, newCookie http.Cookie, newUUID string) bool {
    encodedCookie,encodeErr := json.Marshal(newCookie)
    if encodeErr != nil {
        fmt.Println("Error JSON encoding cookie map")

        if printToFile == 1 {
	    defer Log.Flush()
	    Log.Error("Error JSON encoding cookie map/r/n")
	    return false
	}
    }

    getTimeout := time.Duration((time.Second / 1000) * time.Duration(*timeout))
    client := http.Client {
	Timeout: getTimeout,
    }
    getURL := "http://" + *authhost + strconv.Itoa(*authport) + "/set"
    urlValues := url.Values{}
    urlValues.Set("cookie", string(encodedCookie[:]))
    urlValues.Set("name", newUUID)
    authResp, err := client.PostForm(getURL, urlValues)

    if err != nil {
	fmt.Println("Error running http.Get to confirm cookie set to authserver")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Error("Error running http.Get to confirm cookie set to authserver\r\n")
	}
	return false
    } else if authResp.StatusCode != 200 {
	//fmt.Println("authserver set returned non-200 status code")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Warn("authserver set returned non-200 status code\r\n")
	}
	return false
    }

    return true
}

func authClear (w http.ResponseWriter, r *http.Request, userUUID string) bool {
    getTimeout := time.Duration((time.Second / 1000) * time.Duration(*timeout))
    client := http.Client {
	Timeout: getTimeout,
    }
    getURL := "http://" + *authhost + strconv.Itoa(*authport) + "/clear?cookie=" + userUUID
    authResp, err := client.Get(getURL)
    if err != nil {
	fmt.Println("Error running http.Get to confirm cookie clear from authserver")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Error("Error running http.Get to confirm cookie clear from authserver\r\n")
	}
	return false
    } else if authResp.StatusCode != 200 {
	//fmt.Println("authserver clear returned non-200 status code")
        if printToFile == 1 {
	    defer Log.Flush()
    	    Log.Warn("authserver clear returned non-200 status code\r\n")
	}
	return false
    }

    return true
}