// controllers.go

package controll

import (
	"BackEndGo/src/Pad"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

// controller for requests (methods)
type Controller struct{}

type Controll_Fun interface {
	Get_ID(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	About(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	LoadFile(w http.ResponseWriter, r *http.Request, p httprouter.Params)
}

func NewController() *Controller {
	return &Controller{}
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
	http status:
		200-->everything went fine
		500--> error in json.Marshal

*/

/*
 *
 * Return the file according to pad id
 * */
func (c Controller) LoadFile(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//request
	padRequest := Pad.PadRequest{}
	json.NewDecoder(r.Body).Decode(&padRequest)
	fmt.Println(padRequest)
	//answer
	var pad Pad.Pad
	file, err := ioutil.ReadFile("SavedFiles/" + padRequest.Id)
	if err != nil {
		//return error message
		pad = Pad.Pad{"", "", "File not exist"}
	} else {
		fileAsString := string(file)
		//request in database for name
		db, err := sql.Open("mysql", "root:useruseruser@/onlineEditor")
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()
		stmt, err := db.Prepare("SELECT name FROM filesMetaData WHERE id=?")
		var fileName string
		err = stmt.QueryRow(padRequest.Id).Scan(&fileName)
		pad = Pad.Pad{padRequest.Id, fileName, fileAsString}
	}
	jsonAnswer, err := json.Marshal(pad)
	fmt.Fprintf(w, "%s", jsonAnswer)
	fmt.Println(string(jsonAnswer))
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
