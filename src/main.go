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
	"./model/DataBaseInfo"
	"./model/Pad_info"
	"./model/Requests"
	"github.com/julienschmidt/httprouter"
)

// map for saving possible out-of-order requsts
var Saved_requests []Requests.Editor_req

// var can_write_to_file sync.Mutex
var SavedReq_Mux sync.Mutex

func main() {

	DataBaseInfo.LoadDBInfo()
	fmt.Println(DataBaseInfo.DBLogInString())
	r := httprouter.New()

	// init API
	Requests.Init()
	handleURLS(r)

	// init global vars channel
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
	r.GET("/GetUsers/:id", c.GetLoggedInUsers)
	// 	POST
	// r.POST(<URL1> , <function>)
	r.POST("/PadHistory", c.GetPadHistory)
	r.POST("/NewPad", c.CreateNewPad)
	// ....
	// ...
	// .

	// 	PUT
	// r.PUT(<URL1> , <function>)
	r.PUT("/Edit", c.Upd_PUT)
	// ....
	// ...
	// .

	//	DELETE
	// r.DELETE(<URL1> , <function>)
	// ....

}

/*
	***IMPORTANT***
	Can change edition policy:
		Requests edit notepad contents which is
		a string kept while connection with client is open

		Each requst modifies notepad contents' string
		at the end of a back-end-defined period contents
		are written to disk (? mins). Also Contents are
		written to disk (flush) at timeout/logout.
				-----OR-----
		Can keep contents in mem and save file when
		user asks (more front end work) probably not
		gonna do that.


	This is supposed to act as a go routine
	running in the background receiving requests
	through the in channel,

	saves them and periodically(5sec, not many words can be written hence can be
	lost if server goes down for some reason)
	write into file serverside
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

	//  start accepting requests
	for {
		r, ok := <-Requests.In
		fmt.Println("Received Req:", r)
		if ok {
			SavedReq_Mux.Lock()
			// save request for sorting and serving
			Saved_requests = append(Saved_requests, r)
			SavedReq_Mux.Unlock()
		}
	}
}

// serve the saved requests that have arrived
// actually edit notepad files
// is called every 5sec for each Handle_Requests go routine
func serve_reqs() {

	// empty slice -OR- no requests have arrived
	if len(Saved_requests) == 0 {
		return
	}

	SavedReq_Mux.Lock()

	/* 	sort requests by time they were created so
	editing in files can be done in the right order	*/

	// fmt.Println("Requests before:", Saved_requests)
	sort.Sort(Requests.Oldest_First(Saved_requests))
	// fmt.Println("Requests After:", Saved_requests)
	fmt.Println("Serving Reqs:", Saved_requests)

	for _, v := range Saved_requests {
		/*	write into file r.Value at position r.OffsetFrom:
			paste: many chars to specific loc
			inpt: one char to location		*/
		if er := write_to_pad(v.Notepad_ID, v); er != nil {
			fmt.Println("Error at serving request:\n\t", v, "\n\t", er)
		}
		// remove request ( POP )
		Saved_requests = append(Saved_requests[:0],
			Saved_requests[1:]...)
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

	// update pad from map
	if pad, ok := control.PadMap[pad_id]; ok {
		if req.OffsetFrom > uint(len(pad.Value)) || req.OffsetTo > uint(len(pad.Value)) {

			er = errors.New(fmt.Sprintf("Bad request (out of bounds) %v", req))
			return
		}

		pad.Value = pad.Value[:req.OffsetFrom] + req.Val + pad.Value[req.OffsetTo:]
		pad.Updates = append(pad.Updates, Pad.Pad_update{req.Val, req.OffsetFrom, req.OffsetTo})

		// signal that pad needs flushing to disk
		pad.Needs_flushing = true
		control.PadMap[pad_id] = pad
	} else {
		fmt.Println(control.PadMap)
		er = errors.New("Could not find pad:" + pad_id)
	}

	return
}

/*
	For all files that are active (being editted and not timed-out)
	kept in the PadMap, update their file in disk
*/
func write_to_pad_files() (er error) {
	for _, pad := range control.PadMap {
		if er = pad.Update_file(); er != nil {
			fmt.Println("Error updating pad_file contents for ", pad.ID)
			fmt.Println("\t------\n", er, "\t------\n")
		}
	}
	return
}
