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
		~make struct that holds users connected to pad and pad contents
		~keep slice of open notepads
*/

import (
	"./Controller"
	"./model/DataBaseInfo"
	"./model/Requests"
	"errors"
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

	DataBaseInfo.LoadDBInfo()

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
	r.DELETE("/Edit", c.Upd_DLT)
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
func Handle_Requests() {
	var next_upd, next_serve time.Time
	// run while server is UP

	// update_Files every 1 min
	next_upd = time.Now().Add(addDur(0, 1, 0))

	// serve reqs every 5 secs
	next_serve = time.Now().Add(addDur(0, 0, 5))
	for {
		r, ok := <-Requests.In
		if ok {
			SavedReq_Mux.Lock()
			Saved_requests = append(Saved_requests, r)
			SavedReq_Mux.Unlock()
			// right request has arrived
		}

		/*
			Timer to update files after 1 min
		*/

		if next_serve.After(time.Now()) {
			serve_reqs()
			// next requests serve will happen after 5 s
			next_serve = time.Now().Add(addDur(0, 0, 5))
		}

		if next_upd.After(time.Now()) {
			update_Files()
			// next update will come after 1 min
			next_upd = time.Now().Add(addDur(0, 1, 0))
		}

	}
}

// serve the saved requests that have arrived
// actually edit notepad files
// is called every 5sec for each Handle_Requests go routine
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
	for i, v := range Saved_requests {
		switch v.Req_type {
		case Requests.Dlt:
			/*
				Delete from file:
					delete contents of file from
					r.OffsetFrom till r.OffsetTo

				possibility to call routines for faster
				response???

				TODO:
					~handle errors
			*/
			if er := doDelete(v.Notepad_ID, v); er == nil {
				// remove served request
				Saved_requests = append(Saved_requests[:i],
					Saved_requests[i+1:]...)
			} else {
				fmt.Println("Error at serving requests:\n\t", er)
			}
		case Requests.Wr:
			/*
				write into file r.Value at position r.OffsetFrom:
					paste: many chars to specific loc
					inpt: one char to location


				TODO:
					~handle errors
			*/
			if er := doWrite(v.Notepad_ID, v); er == nil {
				// remove served requesr
				Saved_requests = append(Saved_requests[:i],
					Saved_requests[i+1:]...)
			} else {
				fmt.Println("Error at serving requests:\n\t", er)
			}

		default:
			fmt.Println("Unknown req, in serve reqs: ", v.Req_type)
		} /*End of Switch*/
	} /*End of for*/
	SavedReq_Mux.Unlock()
}

/*
	Parse request received and write value in pad's value that is
	that is kept at global PadMap (controllers.go)
*/
func doWrite(pad_id string, req Requests.Editor_req) (er error) {
	if pad, ok := control.PadMap[pad_id]; ok {
		//change value of pad in  mem
		pad.Value = pad.Value[:req.OffsetFrom] + req.Val + pad.Value[req.OffsetFrom:]
		pad.Need_update()
		control.PadMap[pad_id] = pad
	} else {
		er = errors.New("Could not find ID:" + pad_id)
	}

	return
}

/*
	Parse request received and delete from pad that is
	that is kept at global PadMap (controllers.go)
	the chars requested
*/
func doDelete(pad_id string, req Requests.Editor_req) (er error) {
	if pad, ok := control.PadMap[pad_id]; ok {
		pad.Value = pad.Value[:req.OffsetFrom] + pad.Value[req.OffsetTo:]
		pad.Need_update()
		control.PadMap[pad_id] = pad
	} else {
		er = errors.New("Could not find ID:" + pad_id)
	}

	return
}

func addDur(h, m, s int) time.Duration {
	return time.Hour*time.Duration(h) + time.Minute*time.Duration(m) +
		time.Second*time.Duration(s)
}

/*
	For all files that are active (being editted and not timedout)
	kept in the PadMap update the file they are referring to
*/
func update_Files() (er error){
	for _, pad := range control.PadMap{
		if er = pad.Update_file(); er!=nil{
				fmt.Println("Error updating pad_file contents for ", pad.ID)
				fmt.Println("\t------\n", er, "\t------\n")
				break
		}
	}
	return
}
