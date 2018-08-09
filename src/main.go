// main.go

package main

/*
	import folders that have func implementetions
	for requests
*/

/*
	***IMPORTANT***
	TODO:
		~define better API endponts
			what requests are expected???

		~need to keep content of notepad in file	DONE
		~need to setup sql for metadata	DONE
		~define all model structs
		~make struct that holds users connected to pad and pad contents
		~keep map of open notepads	DONE
		~check that request offsets are in document contents bound		**IMPOTANT**
		~Change request handling to only Write requests			**IMPOTANT**
*/

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
	"./Controller"
	"./model/Users"
	"./model/DataBaseInfo"
	"./model/Pad_info"
	"./model/Requests"
	"github.com/julienschmidt/httprouter"
)

// map for saving possible out-of-order requests
var Saved_requests []Requests.Editor_req

// var can_write_to_file sync.Mutex
var SavedReq_Mux sync.Mutex

func main() {

	// load info from config file
	DataBaseInfo.LoadDBInfo()
	// fmt.Println(DataBaseInfo.DBLogInString())

	r := httprouter.New()

	// init global vars channel
	Requests.Init()
	// init API
	handleURLS(r)

	/*
		possibly multiple go routines can be created for faster
		request-handling/writing to file, if so can_write_to_file
		mutex will come in handy and a mutex for slice access
		is possibly necessary
	*/
	go Init_Editor()

	//close channel when done
	defer Requests.CloseChannel()

	// fire up server
	// check for available ports??
	fmt.Println("Listening to port 8000...")
	http.ListenAndServe(":8000", r)
}

/*
	Assign functions to handle requests
	~maybe URLS need to change??
*/
func handleURLS(r *httprouter.Router) {
	c := control.NewController()

	// 	GET
	r.GET("/OnlineEditor/About", c.About)
	r.GET("/LoadPad/:id", c.LoadPad)
	// r.GET("/GetUsers/:id", c.GetConnectedUsers)
	
	// 	POST
	r.POST("/PadHistory", c.GetPadHistory)
	r.POST("/NewPad", c.CreateNewPad)
	r.POST("/RenameFile", c.RenameFile)
	r.POST("/EmptyFile", c.EmptyDocument)
	
	// 	PUT
	r.PUT("/Edit", c.Upd_PUT)
	
	//	DELETE
	r.DELETE("/DeleteFile", c.DeleteFile)
	
}

/*
	Can change edition policy:
		Requests edit notepad contents which is
		a string kept while connection with client is open

		Each request modifies notepad contents' string
		at the end of a back-end-defined period.
		Contents are written to disk every 30secs.
				-----OR-----
		 	At timeout/logout.

	This is supposed to act as a go routine
	running in the background receiving requests
	through the in channel,

	saves them and periodically(5sec, not many words can be written hence can be
	lost/received out-of-order if server goes down for some reason)
*/
func Init_Editor() {

	// init timer for updating pad files in disk every 30 secs
	go func() {
		writeFiles := time.NewTicker(30 * time.Second)
		for _ = range writeFiles.C {
			write_to_pad_files()
		}
	}()

	// init timer for serving reqs (write in pad value) every 5 secs
	go func() {
		serve_request := time.NewTicker(5 * time.Second)
		for _ = range serve_request.C {
			serve_reqs()
		}
	}()

	// timeoutimplementation
	go func(){
		checkTimeout := time.NewTicker(1 * time.Minute)
		for _ = range checkTimeout.C {
			Users.CleanInactiveUsers()
		} 
	}()

	//  start accepting requests
	for {
		r, ok := <-Requests.In
		fmt.Println("Received Req:", r)

		if ok {
			SavedReq_Mux.Lock()

			// save request for later sorting and serving
			Saved_requests = append(Saved_requests, r)
			SavedReq_Mux.Unlock()
		}
	}
}

// Serve the saved requests that have arrived
// is called every 5sec for each Handle_Requests go routine
func serve_reqs() {

	SavedReq_Mux.Lock()
	// emptied slice -OR- no requests have arrived
	if len(Saved_requests) == 0 {
		SavedReq_Mux.Unlock()
		return
	}


	/* 	sort requests by the time they were created
		to handle posible out-of-order requests	*/
	fmt.Println("Requests before:", Saved_requests)
	sort.Sort(Requests.Oldest_First(Saved_requests))
	fmt.Println("Requests After:", Saved_requests)

	fmt.Println("Serving Reqs:")
	for _,val := range Saved_requests{
		fmt.Println(val)
	}

	for _, v := range Saved_requests {
		fmt.Println("Serving: ", v)
		if er := write_to_pad(v.Notepad_ID, v); er != nil {
			fmt.Println("Error at serving request:\n\t", v, "\n\t", er)
		}

		// remove request ( POP )
		Saved_requests = Saved_requests[1:]
	}
	SavedReq_Mux.Unlock()
}

/*
	Parse request received and update pad that is kept at
	global PadMap (controllers.go)

	Checks are made so that request position are not out of
	pad contents' bounds
*/
func write_to_pad(pad_id string, req Requests.Editor_req) (er error) {

	// get pad from map
	if pad, ok := Pad.PadMap[pad_id]; ok {
		if req.OffsetFrom > uint(len(pad.Value)) || req.OffsetTo > uint(len(pad.Value)) {
			fmt.Println("Value:",pad.Value, " Req_From:",  req.OffsetFrom,
					" Req_To:", req.OffsetTo, "\n Bound:",len(pad.Value))
			er = errors.New(fmt.Sprintf("Bad request (out of bounds) %v", req))
			return
		}

		pad.Value = pad.Value[:req.OffsetFrom] + req.Val + pad.Value[req.OffsetTo:]
		
		// add update to pad to inform client when it asks
		pad.Updates = append(pad.Updates, Pad.Pad_update{req.Val, req.OffsetFrom, req.OffsetTo})

		// signal that pad needs flushing to disk
		pad.Needs_flushing = true

		// update map
		Pad.PadMap[pad_id] = pad
	} else {
		fmt.Println(Pad.PadMap)
		er = errors.New("Could not find pad:" + pad_id)
	}

	return
}

/*
	For all files that are active (being editted and not timed-out)
	kept in the PadMap, update their file in disk
*/
func write_to_pad_files() (er error) {
	for _, pad := range Pad.PadMap {
		if er = pad.Update_file(); er != nil {
			fmt.Println("Error updating pad_file contents for ", pad.ID)
			fmt.Println("\t------\n", er, "\t------\n")
		}
	}
	return
}
