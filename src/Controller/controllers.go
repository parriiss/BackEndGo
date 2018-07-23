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
	"../model/Pad_info"
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


// struct for parsing client request
type PadRequest struct{        Id string `json:"id"`	}
/*
 * Return the info and value of padFile according to pad id
 */
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
	var pad Pad.Pad_info
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
		pad = Pad.Pad_info{padRequest.Id, fileName, fileAsString}
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
		pad = Pad.Pad_info{"", "", errorMessage}
		w.WriteHeader(500)
	}
	jsonAnswer, err := json.Marshal(pad)

	fmt.Fprintf(w, "%s", jsonAnswer)
}

/*
	return the history of pad according to	pad id
	TODO:
		~check if file exist in global map
			if not return 500 error
*/
func (c Controller) GetPadHistory(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	errorFlag := false
	
	//request
	//take the pad id
	padRequest := PadRequest{}
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


var pad_num = 0

var PadMap = make(map[string]*Pad.Pad_info)


/*
	TODO:
		~Add documentation, return values and description 
*/
func (c Controller) CreateNewPad (w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type","application/json")

	// fmt.Fprint(w,"CreateNewPad\n")

	s:=strconv.Itoa(pad_num)
	s="Newpad"+s

	str, er := generate_Pad_Name()

	if er != nil {
		// return internal error status at client
		// regen.Generate returned error
		w.WriteHeader(500)
		fmt.Println("----------\n", er)
		return
	}

	// increment pad name int for next pad creation
	pad_num++
	

	PadMap[str] = &Pad.Pad_info{str, s, "", }

	f:="./SavedFiles/"+str+".txt"
	_,er = os.Create(f)
	if er != nil {
		// could not create file in server
		w.WriteHeader(500)
		fmt.Println("----------\n", er)
		
		// delete from map pad that could not create 
		// and reduce counter for name creation
		pad_num--
		delete(PadMap , str)
		return
	}

	/*
		insertion to pad must be last thing that is done at 
		pad creation because if an error occurs after
		another cpnnection to db must be made so that
		record of pad must be deleted 
	*/
	er = insert_padID_to_db(str, s)
	if er != nil {
		// return internal error status at client
		// db.Open or db.Prepare or Exec returned error
		// couldn't insert to database
		
		// delete from map pad that could not insert to db 
		// and reduce counter for name creation
		pad_num--
		delete(PadMap , str)

		// delete file at server of pad that could not insert to db
		if er2 := os.Remove(f); er2 != nil{
			fmt.Println("----------\n", er2)
		}

		w.WriteHeader(500)
		fmt.Println("----------\n", er)
		return
	}

	
	// pad created, return created status at client
	w.WriteHeader(204)

	// return to client pad that was created
	uj := json.NewEncoder(w).Encode(PadMap[str])
	fmt.Fprintf(w,"%s", uj)

	// print_padMap()	
}

func print_padMap(){
	for k, v := range PadMap {
			fmt.Printf("key[%s] value[%s]\n", k, v)
	}
}

/*
	Generates new unique id for pad
*/
func generate_Pad_Name() (str string, er error){
	for {
		str, er = reggen.Generate("[a-f0-9]{16}", 16)
		if er != nil {
			// return error
			return
		}
		if _,ok :=PadMap[str]; !ok {
			// new pad ID created
			return
		}

		// created a pad ID that already exists
		// try again
	}
}


/*
	Insert new pad Id to db
*/
func insert_padID_to_db(id ,name string) (er error){
	db, er := sql.Open("mysql","root:root@tcp(127.0.0.1:3306)/onlineEditor")
	defer db.Close()
	
	stmt,er := db.Prepare("INSERT INTO filesMetaData SET id=? , name=?")
	if er != nil{
	   return 
	}
	
	_, er = stmt.Exec(id, name)
	
	return
}

/*
	TODO:
		~Add documentation, return values and description 
*/
func (c Controller) RenameFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	// fmt.Fprint(w,"RenameFile\n")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type","application/json")
	

	var t Pad.Pad_info
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		// could not decode json received 
		// return status bad request
		w.WriteHeader(400)
		return
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Found", val.Name)
		val.Name = t.Name

		if err = update_filename_atDb(t.ID, t.Name); err!=nil{
			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return			
		}

		// update value of map if no error has happened
		PadMap[t.ID] = val
	}else{
		fmt.Println("File %s not found",t.ID)
		w.WriteHeader(400)
		return
	}
}

func update_filename_atDb(padId, newName string) (err error) {
	db, err := sql.Open("mysql","root:root@tcp(localhost:3306)/onlineEditor")
	if err != nil { return }	
        	defer db.Close()

	stmt,err := db.Prepare("UPDATE filesMetaData SET name=? WHERE id=? ")
	if err != nil { return }

	_, err = stmt.Exec(newName, padId)

	return
}

/*
	TODO:
		~Add documentation, return values and description 
*/
func (c Controller) DeleteFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	// fmt.Fprint(w,"DeleteFile\n")

	var t Pad.Pad_info

	
	err = json.NewDecoder(r.Body).Decode(&t)
	defer r.Body.Close()

	if err != nil {

		// bad json from client, could not decode
		// return bad request status 
		w.WriteHeader(400)
		return
	}

	if val,ok :=PadMap[t.ID]; ok {
		
		fmt.Println("Delete", val.Name)
		/*
			TODO:
				~keep a temp file( maybe move original)
					if error happens in next steps so
					that you can go back to and not
					remove file
		*/
		err := os.Remove("./SavedFiles/"+PadMap[t.ID].ID+".txt")
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return			
		}
		
		deletePad_fromDb(t.ID)
   		if err != nil {
			/*
				TODO:
					~if error happens at database connection recover deleted file
   			*/

   			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return			
   		}
		
		// almost impossible for an erro to happen here 
		delete(PadMap,t.ID)
	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func deletePad_fromDb(padID string) (err error) {
	db, err := sql.Open("mysql","root:root@tcp(localhost:3306)/onlineEditor")
	if err != nil { return }
	defer db.Close()
	
	stmt,err := db.Prepare("DELETE FROM filesMetaData where id=? ")
	if err != nil { return }

	_, err = stmt.Exec(padID)
	return
}

/*
	TODO:
		~Add documentation, return values and description 
		~Return better 400 errors to clien so it can know if 
			server could not find file or decode json 
*/
func (c Controller) EmptyDocument(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){

	var t Pad.Pad_info
	err :=  json.NewDecoder(r.Body).Decode(&t)
	defer r.Body.Close()
	if err != nil {
		//  bad request, could not decode json
		w.WriteHeader(400)
		return
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Empty Document : ", val.Name)
		err := os.Truncate("./SavedFiles/"+PadMap[t.ID].ID+".txt", 0)
		if err != nil {

			w.WriteHeader(500)
			fmt.Fprintf(w,"%s",err)
		}
	}else{
		fmt.Println("File %s not found",t.ID)
		//  bad request, could find requested file
		w.WriteHeader(400)

	}

}


