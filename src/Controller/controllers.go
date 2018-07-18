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
	"BackEndGo/src/Pad"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"time"
)

// controller for requests (methods)
type Controller struct{}


type Controll_Fun interface{
	Get_ID(w http.ResponseWriter ,r *http.Request, p httprouter.Params)
	About(w http.ResponseWriter ,r *http.Request, p httprouter.Params)
	Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
  LoadFile(w http.ResponseWriter, r *http.Request, p httprouter.Params)
} 

func NewController() *Controller{
	return  &Controller{}
}

/*
	get ID for Notepad??
	to be implemented
*/
func (c Controller) Get_ID(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {
	// ??
}

/*
	Gets a request from client for the about page
	response json:
		{	Lang	:	"Golang" 	}
	http response header status:
		200-->everything went fine  
		500-->error in json.Marshal

	http status:
		200-->everything went fine
		500--> error in json.Marshal


*/

/*
 *
 * Return the info and value of padFile according to pad id
 * */
func (c Controller) LoadFile(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	errorFlag := false
	errorMessage := ""
	//request
	padRequest := Pad.PadRequest{}
	json.NewDecoder(r.Body).Decode(&padRequest)
	//answer
	var pad Pad.Pad
	file, err := ioutil.ReadFile("SavedFiles/" + padRequest.Id)
	if err != nil {
		errorMessage = "File not exist"
		errorFlag = true
	} else {
		fileAsString := string(file)
		//request in database for name
		db, err := sql.Open("mysql", "root:useruseruser@/onlineEditor")
		if err != nil {
			errorMessage = "error db"
			errorFlag = true
		}
		defer db.Close()
		stmt, err := db.Prepare("SELECT name FROM filesMetaData WHERE id=?")
		if err != nil {
			errorMessage = "error db"
			errorFlag = true
		}
		var fileName string
		err = stmt.QueryRow(padRequest.Id).Scan(&fileName)
		if err != nil {
			errorMessage = "error db"
			errorFlag = true
		}
		pad = Pad.Pad{padRequest.Id, fileName, fileAsString}
		//insert in db info about user started session
		//time format
		logInTime := string(time.Now().Format("2006-01-02 15:04:05"))
		userIp := string(r.RemoteAddr)
		//state=1 :: started session
		state := 1

		stmt, err = db.Prepare("INSERT INTO historyFiles SET ip=?, id=?, time=?, state=?")
		if err != nil {
			errorMessage = "error db"
			errorFlag = true
		}
		_, err = stmt.Exec(userIp, padRequest.Id, logInTime, state)
		if err != nil {
			errorMessage = "error db"
			errorFlag = true
		}
	}
	if errorFlag == true {
		pad = Pad.Pad{"", "", errorMessage}
	}
	jsonAnswer, err := json.Marshal(pad)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", jsonAnswer)
}

func (c Controller) About(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	rj, er := json.Marshal(struct {
		Lang string
	}{
		Lang: "Golang",
	})

	if er != nil {
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
