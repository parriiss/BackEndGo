// controllers.go

package control

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
	"../model/PadHistory"
	"../model/Requests"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"time"
	"os"
	"BackEndGo/src/model/pad_options"
	"strconv"
	"github.com/lucasjones/reggen"
	
)

// controller for requests (methods)
type Controller struct{}

type Control_Fun interface {
	About(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	Upd_DLT(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	LoadFile(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Get_ID(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetPadHistory(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateNewPad(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	RenameFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	DeleteFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	EmptyDocument(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
}

var (
	fileInfo *os.FileInfo
	err      error
)

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
	padRequest := PadRequest{}
	json.NewDecoder(r.Body).Decode(&padRequest)
	//answer
	var pad Pad
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
		pad = Pad{padRequest.Id, fileName, fileAsString}
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
	w.WriteHeader(200)
	if errorFlag == true {
		pad = Pad{"", "", errorMessage}
		w.WriteHeader(500)
	}
	jsonAnswer, err := json.Marshal(pad)

	fmt.Fprintf(w, "%s", jsonAnswer)
}

/*
return the history of pad according to
pad id
//TODO: check if file exist in global map
if not return 500 error
*/
func (c Controller) GetPadHistory(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	errorFlag := false
	//request
	//take the pad id
	padRequest := model.PadRequest{}
	json.NewDecoder(r.Body).Decode(&padRequest)
	//answer
	//values of table historyFiles in DB
	var (
		id    string
		state int
		time  string
		ip    string
		//slice with all history values
		history []PadHistory.PadHistory
	)

	//TODO check if exist the pad with this id

	//connect to db
	db, err := sql.Open("mysql", "root:useruseruser@/onlineEditor")
	if err != nil {
		errorFlag = true
	}
	//close th db
	defer db.Close()

	//query to db to take the history of pad
	sqlStatement := `SELECT * FROM historyFiles WHERE id=?`
	rows, err := db.Query(sqlStatement, padRequest.Id)
	//iterate the results from query
	for rows.Next() {
		//read the values per row
		err = rows.Scan(&ip, &id, &time, &state)
		if err != nil {
			errorFlag = true
		}
		//insert them to the slice
		historyToInsert := PadHistory.PadHistory{ip, state, time}
		history = append([]PadHistory.PadHistory{historyToInsert}, history...)
	}
	w.WriteHeader(200)
	if errorFlag == true {
		w.WriteHeader(500)
	}
	jsonAnswer, err := json.Marshal(history)
	fmt.Fprintf(w, "%s", jsonAnswer)
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
	if c_req.OffsetTo > 0 {
		t = Requests.Ins
	}

	s_req = Requests.Editor_req{
		Req_date:   c_req.Req_date,
		Req_type:   t,
		Val:        c_req.Val,
		OffsetFrom: c_req.OffsetFrom,
		OffsetTo:   c_req.OffsetTo,
	}

	// 	put req in channel for routine to handle
	Requests.In <- s_req

	w.WriteHeader(202)
}

/*
	Empty Function to handle DELETE request when deletion is happenning at
	Edit page
*/
func (c Controller) Upd_DLT(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

//----------------------Options for Pad---NewPad(StorePad),Delete,Rename,EmptyDocument---------------------------
type Pad struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Value string `json:"value"`
}

func NewPad() *Pad{
	return  &Pad{}
}

type PadRequest struct{
        Id string `json:"id"`
}



var i=0

var PadMap=make(map[string]*Pad)


func (c Controller) CreateNewPad (int)(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){

	fmt.Fprint(w,"CreateNewPad\n")
   db, err := sql.Open("mysql",
                "root:root@tcp(127.0.0.1:3306)/onlineEditor")
	w.WriteHeader(201)
        if err != nil {
                //panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
		w.WriteHeader(500)
        }
        defer db.Close()


	s:=strconv.Itoa(i)
	s="Newpad"+s

	for{
		str, err2 := reggen.Generate("[a-f0-9]{16}", 16)
		if err2 != nil {
			//panic(err2)
			w.WriteHeader(500)
		}
		
		if val,ok :=PadMap[str]; ok {
			fmt.Println("	Found",val.Name)

		}else{
			PadMap[str]=&Pad{
				str,
				s,
				"",
			}
			f:="./SavedFiles/"+str+".txt"
			os.Create(f)

			stmt,err := db.Prepare("INSERT INTO filesMetaData SET id=? , name=?")
			if err != nil {
			//	panic(err)
			w.WriteHeader(500)
		}
			_, err = stmt.Exec(str, s)
			if err != nil {
					//panic(err)
					 w.WriteHeader(500)

			}

			i=i+1
			break;
			}
return str;

   }






	for k, v := range PadMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}

	fmt.Printf("----------\n")
if err!=nil{
fmt.Fprintf(w,"%s",err)
}else{
fmt.Fprintf(w,"")
}

}
func (c Controller) RenameFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"RenameFile\n")
db, err := sql.Open("mysql",
                "root:root@tcp(localhost:3306)/onlineEditor")
w.WriteHeader(200)
        if err != nil {
                //panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
		w.WriteHeader(500)
        }
        defer db.Close()


	decoder := json.NewDecoder(r.Body)
	var t Pad
	err = decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(500)
		//panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Found", val.Name)
		PadMap[t.ID].Name=t.Name
 stmt,err := db.Prepare("UPDATE filesMetaData SET name=? WHERE id=? ")
   if err != nil {
	w.WriteHeader(500)
       //panic(err)
   }
		_, err = stmt.Exec(t.Name, t.ID)
			if err != nil {
			//		panic(err)
					 w.WriteHeader(500)

			}

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func (c Controller) DeleteFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var t Pad
    	w.WriteHeader(200)
fmt.Fprint(w,"DeleteFile\n")
db, err := sql.Open("mysql",
                "root:root@tcp(localhost:3306)/onlineEditor")



	err = decoder.Decode(&t)
	if err != nil {
		w.WriteHeader(500)
		//panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
			fmt.Println("Delete", val.Name)
		err := os.Remove("./SavedFiles/"+PadMap[t.ID].ID+".txt")
		if err != nil {
			w.WriteHeader(500)
			//log.Fatal(err)
		}
		 stmt,err := db.Prepare("DELETE FROM filesMetaData where id=? ")
   if err != nil {
       w.WriteHeader(500)
	//panic(err)
   }
		_, err = stmt.Exec(t.ID)
			if err != nil {
			//		panic(err)
					 w.WriteHeader(500)

			}
		delete(PadMap,t.ID)
		

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func (c Controller) EmptyDocument(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var t Pad
	w.WriteHeader(200)
	err := decoder.Decode(&t)
	if err != nil {
		//panic(err)
		w.WriteHeader(500)
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Empty Document : ", val.Name)
		err := os.Truncate("./SavedFiles/"+PadMap[t.ID].ID+".txt", 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w,"%s",err)
			w.WriteHeader(500)
		}
	}else{
		fmt.Println("File %s not found",t.ID)
	}

}


