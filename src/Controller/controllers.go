// controllers.go

package control

import (
	//"../model/DataBaseInfo"
	"../model/DataBaseInfo"
	"../model/PadHistory"
	"../model/Pad_info"
	"../model/Requests"
	"../model/Users"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"github.com/lucasjones/reggen"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// controller for requests (methods)
type Controller struct{}

type Control_Fun interface {
	About(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	LoadPad(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetConnectedUsers(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	GetPadHistory(w http.ResponseWriter, r *http.Request, p httprouter.Params)
	CreateNewPad(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	RenameFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	DeleteFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
	EmptyDocument(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
}

var (
	fileInfo *os.FileInfo
	err      error
)

func NewController() *Controller {
	return &Controller{}
}

// struct for parsing client request
type PadRequest struct {
	Id string `json:"id"`
}

/*
	function that takes a pad id and return the content of the current pad
	If the pad dont exist it return an empty string and an error
	if the pad exist but it is empty return an empty string and a nil error
	otherwise return nil error and the content
*/
func (c Controller) LoadPadFromFile(padId string) (string, error) {
	file, err := ioutil.ReadFile("SavedFiles/" + padId + ".txt")
	if err != nil {
		return "", err
	}
	fileAsString := string(file)
	return fileAsString, err
}

/*
  'GET' function that returns
  the info and value of padFile according
  to pad id

  Respond back with:
        StatusCode:200 Success,Ok
        StatusCode:500 Server Error(Error in DB)
        StatusCode:404 Could not find pad or pad-file

*/
func (c Controller) LoadPad(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var (
		err                    error
		fileAsString, fileName string
		pad                    *Pad.Pad_info
		padFound               bool
	)

	//request
	requestID := p.ByName("id")
	u := Users.User{r.RemoteAddr, time.Now()}

	/*
		if pad exists in padmap
			load it
		else
			try to load file
			if pad id not found
				return 404
	*/
	Pad.PadLock.Lock()
	if pad = Pad.PadMap[requestID]; pad != nil {
		fileAsString = pad.Value
		padFound = true
	} else if fileAsString, err = c.LoadPadFromFile(requestID); err != nil {
		w.WriteHeader(404)
		Pad.PadLock.Unlock()
		return
	}

	/* if pad found in map check if a new user has connected */
	newuser := true
	if padFound {
		for _, u := range pad.Users {
			if u.Address == r.RemoteAddr {
				newuser = false
				break
			}
		}
	}

	//database for name / session start
	db := &sql.DB{}

	/* if pad is not in map
		get name from DB
	else if this is a new user
		add user to pad
	*/
	if !padFound {
		if db, err = sql.Open("mysql", DataBaseInfo.DBLogInString()); err != nil {
			w.WriteHeader(500)
			return
		}
		defer db.Close()

		stmt, err := db.Prepare("SELECT name FROM filesMetaData WHERE id=?")
		if err != nil {
			w.WriteHeader(500)
			return
		}

		if err = stmt.QueryRow(requestID).Scan(&fileName); err != nil {
			w.WriteHeader(500) //at db error
			return
		}

		start_session(requestID, r.RemoteAddr, db)
	} else if newuser {
		// pad is found in map AND a new user is connected signal start session
		Pad.PadLock.Lock()
		Pad.InsertUserIp(u.Address, pad.ID)
		Pad.PadLock.Unlock()

		if db, err = sql.Open("mysql", DataBaseInfo.DBLogInString()); err != nil {
			w.WriteHeader(500)
			return
		}
		defer db.Close()

		start_session(pad.ID, r.RemoteAddr, db)
	}

	//keep the pad in the global pad map if not there
	if !padFound {
		pad = &Pad.Pad_info{requestID,
			fileName,
			fileAsString,
			nil,
			false,
			[]Users.User{u},
		}

		Pad.PadLock.Lock()
		Pad.PadMap[requestID] = pad
		Pad.PadLock.Unlock()
	}

	//add the user to the global map logedInUsers
	w.WriteHeader(200)

	jsonAnswer, err := json.Marshal(Pad.PadResponse{
		ID:    pad.ID,
		Name:  pad.Name,
		Value: fileAsString,
		Users: pad.Users,
	})

	fmt.Fprintf(w, "%s", jsonAnswer)
}

func start_session(id, ip string, db *sql.DB) (err error) {
	//time format

	//insert in db info about user started session
	stmt, err := db.Prepare("INSERT INTO historyFiles SET ip=?, id=?, time=?, state=?")
	if err != nil {
		return
	}

	logInTime := string(time.Now().Format("2006-01-02 15:04:05"))
	//state : 1 --> started session
	_, err = stmt.Exec(ip, id, logInTime, 1)

	return
}

/*
	Get Method to return according to padId all
	the users from global map logedinusers that they are edititng the pad
	in case of an error (no one edit the pad or the pad dont exist)
	return an empty array
*/
func (c Controller) GetConnectedUsers(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {

	padId := p.ByName("id")
	users := Pad.GetConnectedUsers(padId)
	jsonAnswer, err := json.Marshal(users)
	if err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	}
}

/*
	return the history of pad according to	pad id
*/
func (c Controller) GetPadHistory(w http.ResponseWriter,
	r *http.Request, p httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	errorFlag := false
	errorMessage := ""
	//values of table historyFiles in DB
	var (
		id    string
		state int
		time  string
		ip    string
		//slice with all history values
		history []PadHistory.PadHistory
	)
	//request
	//take the pad id
	padRequest := PadRequest{}
	json.NewDecoder(r.Body).Decode(&padRequest)
	//error in json from request
	if padRequest.Id == "" {
		errorFlag = true
		errorMessage = "error in json from request"
	} else {
		//answer
		//connect to db
		db, err := sql.Open("mysql", DataBaseInfo.DBLogInString())
		if err != nil {
			errorFlag = true
			errorMessage = "cant connect to db"
		} else {
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
			if len(history) == 0 {
				errorFlag = true
				errorMessage = "no history for this pad"
			}
		}
	}
	if errorFlag == true {
		w.WriteHeader(500)
		fmt.Fprintf(w, errorMessage)
		return
	}
	w.WriteHeader(200)
	jsonAnswer, err := json.Marshal(history)
	if err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	}

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
-Gets a request to create a new NotePad
-Respond back with:
	StatusCode:200 Success,Ok
	StatusCode:500 Server Error(Fail to create a file,or generate a new ID)

*/
func (c Controller) CreateNewPad(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	fmt.Println(DataBaseInfo.FolderDir.FilesDir)
	s := "Newpad"
	str, er := generate_Pad_Name()

	if er != nil {
		// return internal error status at client
		// regen.Generate returned error
		w.WriteHeader(500)
		fmt.Println("----------\n", er)
		return
	}

	u := Users.User{r.RemoteAddr, time.Now()}
	p := &Pad.Pad_info{str, s, "", nil, false, []Users.User{u}}
	//  ID, filename , contentsOfPad , Updates , Needs_flushing , ConnectedUsers
	Pad.PadMap[str] = p

	if _, err := os.Stat(DataBaseInfo.FolderDir.FilesDir); os.IsNotExist(err) {
		os.Mkdir(DataBaseInfo.FolderDir.FilesDir, 0755)
	}
	f := DataBaseInfo.FolderDir.FilesDir + str + ".txt"
	_, er = os.Create(f)
	if er != nil {
		// could not create file in server
		w.WriteHeader(500)
		fmt.Println("----------\n", er)

		// delete from map pad that could not create
		delete(Pad.PadMap, str)
		return
	}

	/*
		insertion to pad must be last thing that is done at
		pad creation because otherwise if an error occurs after
		another connection to db must be made so that
		record of pad must be deleted
	*/
	er = insert_padID_to_db(str, s, r.Host)
	if er != nil {
		// return internal error status at client
		// db.Open or db.Prepare or Exec returned error
		// couldn't insert to database

		// delete from map pad that could not insert to db
		// and reduce counter for name creation
		delete(Pad.PadMap, str)

		// delete file at server of pad that could not insert to db
		if er2 := os.Remove(f); er2 != nil {
			fmt.Println("----------\n", er2)
		}
		w.WriteHeader(500)
		fmt.Println("----------\n", er)
		return
	}

	// pad created, return created status at client
	w.WriteHeader(200)
	pr := Pad.PadResponse{
		ID:    p.ID,
		Name:  p.Name,
		Value: p.Value,
		Users: p.Users,
	}
	// return to client pad that was created
	if jsonAnswer, err := json.Marshal(pr); err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	} else {
		fmt.Println("Error in marshal", er, pr)
	}

	// print_padMap()
}

// For Debugging
func print_padMap() {
	for k, v := range Pad.PadMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}
}

/*
	Generates new unique id for pad
*/
func generate_Pad_Name() (str string, er error) {
	for {
		str, er = reggen.Generate("[a-f0-9]{16}", 16)
		if er != nil {
			// return error
			return
		}
		if _, ok := Pad.PadMap[str]; !ok {
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
func insert_padID_to_db(id, name, ip string) (er error) {
	db, er := sql.Open("mysql", DataBaseInfo.DBLogInString())
	defer db.Close()

	stmt, er := db.Prepare("INSERT INTO filesMetaData SET id=? , name=?")
	if er != nil {
		return
	}

	_, er = stmt.Exec(id, name)
	logInTime := string(time.Now().Format("2006-01-02 15:04:05"))

	stmt, err := db.Prepare("INSERT INTO historyFiles SET ip=?, id=?, time=?, state=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(ip, id, logInTime, 1)
	if err != nil {
		return
	}

	return
}

/*
Gets a Request to rename a file with specific ID
-Request: JSON->id

-Respond back with:
        StatusCode:200 Success,Ok
        StatusCode:500 Server Error(Fail to create a file,or generate a new ID)
		StatusCode:400 Could not decode JSON
		StatusCode:404 Could not find Pad
*/
func (c Controller) RenameFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// fmt.Fprint(w,"RenameFile\n")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var t Pad.Pad_info
	err = json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		// could not decode json received
		// return status bad request
		w.WriteHeader(400)
		return
	}
	val := Pad.PadMap[t.ID]

	if val, ok := Pad.PadMap[t.ID]; ok {
		fmt.Println("Found", val.Name)
		val.Name = t.Name

		if err = update_filename_atDb(t.ID, t.Name); err != nil {
			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return
		}

		// update value of map if no error has happened
		Pad.PadMap[t.ID] = val
	} else {
		fmt.Println("Pad %s not found", t.ID)
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
	pr := Pad.PadResponse{
		ID:    val.ID,
		Name:  val.Name,
		Value: val.Value,
		Users: val.Users,
	}
	// return to client pad that was created
	if jsonAnswer, err := json.Marshal(pr); err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	} else {
		fmt.Println("Error in marshal", err, pr)
	}
}

func update_filename_atDb(padId, newName string) (err error) {
	db, err := sql.Open("mysql", DataBaseInfo.DBLogInString())
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE filesMetaData SET name=? WHERE id=? ")
	if err != nil {
		return
	}

	_, err = stmt.Exec(newName, padId)

	return
}

/*
Gets a Request to delete a file with specific ID
-Request: JSON->id

-Respond back with:
        StatusCode:200 Success,Ok
        StatusCode:500 Server Error(Fail to remove the requested file locally)
        StatusCode:400 Could not decode JSON
*/
func (c Controller) DeleteFile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	val := Pad.PadMap[t.ID]

	if val, ok := Pad.PadMap[t.ID]; ok {

		fmt.Println("Delete", val.Name)
		/*

		   ~keep a temp file( maybe move original)
		   if error happens in next steps so
		    that you can go back to and not
		     remove file
		*/

		recPath := "./" + t.ID + "-Backup" + ".txt"

		originalPath := DataBaseInfo.FolderDir.FilesDir + Pad.PadMap[t.ID].ID + ".txt"
		CreateBackupFile(originalPath, recPath)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return
		}

		err = os.Remove(originalPath)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		err = deletePad_fromDb(t.ID)
		if err != nil {
			/*
			   ~if error happens at database connection recover deleted file
			*/

			err := os.Rename(recPath, originalPath)
			if err != nil {
				w.WriteHeader(500)
				return
			}

			w.WriteHeader(500)
			fmt.Println("----------\n", err)
			return
		}

		err = os.Remove(recPath)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		// almost impossible for an error to happen here
		delete(Pad.PadMap, t.ID)
	} else {
		fmt.Println("File %s not found", t.ID)
		w.WriteHeader(404)
	}

	w.WriteHeader(200)
	pr := Pad.PadResponse{
		ID:    val.ID,
		Name:  val.Name,
		Value: val.Value,
		Users: val.Users,
	}
	// return to client pad that was created
	if jsonAnswer, err := json.Marshal(pr); err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	} else {
		fmt.Println("Error in marshal", err, pr)
	}
}

/*
Gets the original path of a file and creates a new file wih the contents of the original one as backup
 Returns err if occurs one
*/
func CreateBackupFile(originalPath string, backupPath string) (err error) {

	newFile, err := os.Create(backupPath)
	if err != nil {

		return
	}
	defer newFile.Close()
	originalFile, err := os.Open(originalPath)
	if err != nil {

		return
	}

	bytesWritten, err := io.Copy(newFile, originalFile)
	if err != nil {

		return
	}
	fmt.Println("Copied %d bytes", bytesWritten)
	err = newFile.Sync()
	if err != nil {

		return
	}

	return
}

func deletePad_fromDb(padID string) (err error) {
	db, err := sql.Open("mysql", DataBaseInfo.DBLogInString())
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM filesMetaData where id=? ")
	if err != nil {
		return
	}

	_, err = stmt.Exec(padID)

	stmt, err = db.Prepare("DELETE  FROM historyFiles WHERE id=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(padID)
	if err != nil {
		return
	}

	return
}

/*
Gets a Request to empty the contents of a file with specific ID
-Request: JSON->id

-Respond back with:
        StatusCode:200 Success,Ok
        StatusCode:500 Server Error(Fail to truncate the requested file)
        StatusCode:400 Could not decode JSON
        StatusCode:404 Could not find pad or pad-file
*/
func (c Controller) EmptyDocument(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var t Pad.Pad_info
	err := json.NewDecoder(r.Body).Decode(&t)
	defer r.Body.Close()
	if err != nil {
		//  bad request, could not decode json
		w.WriteHeader(400)
		return
	}
	val := Pad.PadMap[t.ID]

	if val, ok := Pad.PadMap[t.ID]; ok {
		fmt.Println("Empty Document : ", val.Name)

		err := os.Truncate(DataBaseInfo.FolderDir.FilesDir+Pad.PadMap[t.ID].ID+".txt", 0)
		if os.IsNotExist(err) {
			w.WriteHeader(404)
			return
		}

		Pad.PadMap[t.ID].Value = ""
	} else {
		fmt.Println("File %s not found", t.ID)
		//  bad request, could find requested file
		w.WriteHeader(404)

		return
	}
	w.WriteHeader(200)

	pr := Pad.PadResponse{
		ID:    val.ID,
		Name:  val.Name,
		Value: val.Value,
		Users: val.Users,
	}
	// return to client pad that was created
	if jsonAnswer, err := json.Marshal(pr); err == nil {
		fmt.Fprintf(w, "%s", jsonAnswer)
	} else {
		fmt.Println("Error in marshal", err, pr)
	}

}

/*
	Pass the incoming request to channel in order to save it
	and serve later all saved requests in the order they were
	created (solve case of out-of-order arrival )

	http response header status:
		202-->request received for processing
			not yet served
		400-->error in json decoding or
			other error checking
*/
func (c Controller) Upd_PUT(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var (
		updates []Pad.Pad_update
		c_req   Requests.Client_Put
		ok      bool
	)
	/*
		possible error json checking here for quick response of
		wrong data to client
	*/
	if er := json.NewDecoder(r.Body).Decode(&c_req); er != nil {
		defer r.Body.Close()
		fmt.Println("Error in decoding json in write Parse_requests\n", er)
		w.WriteHeader(400)
		return
	}
	defer r.Body.Close()

	fmt.Println("Received req for pad:", c_req.Notepad_ID)
	_, ok = Pad.PadMap[c_req.Notepad_ID]
	if !ok {
		fmt.Println("Pad:", c_req.Notepad_ID, " not found")
		// requested pad not found
		w.WriteHeader(404)
		return
	} else if !c_req.Is_update_request {
		// 	put req in channel for routine to handle
		Requests.In <- Requests.Editor_req{
			Timestamp:  c_req.Timestamp,
			Val:        c_req.Val,
			OffsetFrom: c_req.OffsetFrom,
			OffsetTo:   c_req.OffsetTo,
			Notepad_ID: c_req.Notepad_ID,
			UserIp:     r.RemoteAddr,
		}

		KeepAlive(r.RemoteAddr, c_req.Notepad_ID)
		w.WriteHeader(202)
	} else {

		/*
			get the updates that user hasn't been informed
			also cleans up updates for anything that everyone
			has been notified
		*/
		updates = Pad.GetUserUpdates(r.RemoteAddr, c_req.Notepad_ID)
		KeepAlive(r.RemoteAddr, c_req.Notepad_ID)

		// response json
		if rj, er := json.Marshal(updates); er != nil {
			w.WriteHeader(500) // failed to mashal json
		} else {
			fmt.Fprintf(w, "%s", rj)
			w.WriteHeader(200)
		}
	}

}

//  this should be method of a struct that contains the users map and
//  not a seperate function in controller
func KeepAlive(userAddress, padID string) {
	if _, ok := Pad.PadMap[padID]; !ok {
		return
	}

	for _, user := range Pad.PadMap[padID].Users {
		if userAddress == user.Address {
			user.KeepActive()
			fmt.Println("Keeping alive ipAddress:", userAddress)
		}
	}
}
