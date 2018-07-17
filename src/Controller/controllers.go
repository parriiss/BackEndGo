// controllers.go

package controll

import (
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
	http status:
		200-->everything went fine  
		500--> error in json.Marshal
		
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
