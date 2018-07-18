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
	"sync"
	"./Controller"

)


// map for saving possible out-of-order requsts
var Saved_requests []Requests.Editor_req

var can_write_to_file sync.Mutex
var sliceMux sync.Mutex

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


// Assign functions to handle requests
func handleURLS(r *httprouter.Router) {
	c := controll.NewController()

	// 	GET
	r.GET("/OnlineEditor/About", c.About)

	

// 	POST
	r.POST("/LoadFile",c.LoadFile)
// r.POST(<URL1> , <function>)

	// ....
	// ...
	// .

	// 	PUT
	r.PUT("/", c.Upd_PUT)

	// r.PUT(<URL1> , <function>)
	// ....
	// ...
	// .

	//	DELETE
	// r.DELETE(<URL1> , <function>)
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
	// run while server is UP
	for {
		r, ok := <-Requests.In
		if ok {
			Saved_requests = append(Saved_requests , r)
			// right request has arrived
		}
		/*
			***IMPORTANT***
			Here a timer check must be implemented so periodically
			requests (characters) are written to file...
		*/

	}
}

// serve the saved requests that have arrived
// actually edit notepad files
// is called every 5sec for each Handle_Requests (subRoutine)
func serve_reqs() {

	/*
		***IMPORTANT***
		The Saved requests must be sorted in order 
		to write each to file in the correct order
	*/

	// unecessary if only one go Hanlde_Requests is called
	// unlocks access to Saved_Requests for serving
	sliceMux.Lock()
	// if request that I'm expecting is saved
	//  possible to have multiple out of order
	for _, v := range Saved_requests{
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
				fmt.Println("Unknown req, in serve reqs: ",v.Req_type)
		}/*End of Switch*/
	} /*End of for*/
	sliceMux.Unlock()
}  

