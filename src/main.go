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

		~need to keep content of notepad in file
		~need to setup sql for metadata
		~define all model structs
*/

import (
	"./Controller"
	"./model/Requests"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sort"
	"sync"
	"time"
)

// map for saving possible out-of-order requsts
var Saved_requests []Requests.Editor_req

var can_write_to_file sync.Mutex
var SavedReq_Mux sync.Mutex

func main() {

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
	go Handle_Requests()

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

	// 	POST
	// r.POST(<URL1> , <function>)
	r.POST("/LoadFile", c.LoadFile)
	r.POST("/PadHistory", c.GetPadHistory)

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
	r.DELETE("/Edit", c.Upd_DLT)
	// ....

}

/*
	this is supposed to act as a go routine
	running in the background receiving requests
	through the in channel,

	saves them and periodically(5sec, not many words can be written hence can be
	lost if server goes down for some reason)
	write into file serverside
*/
func Handle_Requests() {
	var next_upd time.Time
	// run while server is UP
	for {
		// update files after 5 secs
		next_upd = time.Now().Add(addDur(0, 0, 5))
		r, ok := <-Requests.In
		if ok {
			SavedReq_Mux.Lock()
			Saved_requests = append(Saved_requests, r)
			SavedReq_Mux.Unlock()
			// right request has arrived
		}

		/*
			Timer to update files after 5secs
		*/
		if next_upd.After(time.Now()) {
			// next update will come after 5secs
			next_upd = time.Now().Add(addDur(0, 0, 5))
			serve_reqs()
		}

	}
}

// serve the saved requests that have arrived
// actually edit notepad files
// is called every 5sec for each Handle_Requests (subRoutine)
func serve_reqs() {

	/*
		TODO:
			***IMPORTANT***

			~Changes in file should be boadcasted to every
			implementation of getPadHistory function
			user connected to the editted notepad
			(broadcast AFTER files changed)

	*/

	// unecessary if only one Hanlde_Requests routine is called
	// unlocks access to Saved_Requests for serving
	SavedReq_Mux.Lock()

	/*
			Either:
		 		another serve_reqs called from handle routine
		 			has emptied slice and unlocked mutex
			OR
				no requests have arrived
	*/
	if len(Saved_requests) == 0 {
		SavedReq_Mux.Unlock()
		return
	}

	// sort requests by time they were created so
	// editing in files can be done in the right order
	sort.Sort(Requests.Oldest_First(Saved_requests))

	// if request that I'm expecting is saved
	//  possible to have multiple out of order
	for _, v := range Saved_requests {
		switch v.Req_type {
		case Requests.Ins:
			/*
				Insert into file:
					delete contents of file from
					r.OffsetFrom till r.OffsetTo
					and insert r.Value to
					r.OffsetFrom position
			*/

		case Requests.Dlt:
			/*
				Delete from file:
					delete contents of file from
					r.OffsetFrom till r.OffsetTo
			*/

		case Requests.Wr:
			/*
				write into file r.Value at position r.OffsetFrom:
					paste: many chars to specific loc
					inpt: one char to location
			*/

		default:
			fmt.Println("Unknown req, in serve reqs: ", v.Req_type)
		} /*End of Switch*/
	} /*End of for*/
	SavedReq_Mux.Unlock()
}

func addDur(h, m, s int) time.Duration {
	return time.Hour*time.Duration(h) + time.Minute*time.Duration(m) +
		time.Second*time.Duration(s)
}
