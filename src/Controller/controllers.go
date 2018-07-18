// controllers.go

package controll

/*
	TODO:	
		~implement new notepad functionality

		~implement editting functionality
			~users currently viewing/editting file
			~updating db file

		~ implement settings
			~clear contents of file
			~dlt doc
			~change doc tittle
			~history of all people that viewed/editted doc
*/		

import (
	"../model/Requests"
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

// controller for requests (methods)
type Controller struct{}

type Controll_Fun interface{
	Get_ID(w http.ResponseWriter ,r *http.Request, p httprouter.Params)
	About(w http.ResponseWriter ,r *http.Request, p httprouter.Params)
	Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
} 

func NewController() *Controller{
	return  &Controller{}
}

/*
	get ID for Notepad?? 
	to be implemented
*/	
func (c Controller)  Get_ID(w http.ResponseWriter ,
	r *http.Request, p httprouter.Params){
	// ??
}


/*
	Gets a request from client for the about page
	response json:
		{	Lang	:	"Golang" 	}
	http response header status:
		200-->everything went fine  
		500-->error in json.Marshal
		
*/
func (c Controller) About(w http.ResponseWriter ,
	r *http.Request, p httprouter.Params){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Content-Type" , "application/json")

	rj, er := json.Marshal(struct{
			Lang string
		}{
			Lang : "Golang",
			})
	
	if er != nil{
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", rj)
}


/*
	http response header status:
		202-->request received for processing 
			not yet served
		400-->error in json decoding or 
			other error checking
*/
func (c Controller) Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	c_req := Requests.Client_Put{}
	s_req := Requests.Editor_req{}

	defer r.Body.Close()
	if er := json.NewDecoder(r.Body).Decode(&c_req); er != nil {
		fmt.Println("Error in decoding json in write Parse_requests")
		w.WriteHeader(400)
		return
	}

	/*
		possible error json checking here for quick response of
		wrong data to client
	*/

	t := Requests.Wr
	if c_req.OffsetTo > 0 {	t = Requests.Ins	}

	s_req = Requests.Editor_req{
		Req_date: c_req.Req_date,
		Req_type:   t,
		Val:        c_req.Val,
		OffsetFrom: c_req.OffsetFrom,
		OffsetTo:   c_req.OffsetTo,
	}

	// 	put req in channel for routine to handle
	Requests.In <- s_req

	w.WriteHeader(202)
}