package DataBaseInfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//a struct to store the db info for log in
type DataBaseInfo struct {
	DBName   string `json:"name"`
	Username string `json:"user"`
	Password string `json:"pass"`
}
type DB struct {
	DB DataBaseInfo `json:"DataBase"`
}
type Folder struct {
	FilesDir string `json:"FilesDir"`
}
type ListenPort struct {
	ListeningPort string `json:"ListeningPort"`
}

// a global instance of DataBaseInfo
var DBInfo DB
var FolderDir Folder
var lport ListenPort

/*
	function to read from DBconfigFile
	json file the info of db
*/
func LoadDBInfo() {
	path := "./DBConfigFile"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Could not open ", path, err)
		return
	}
	defer file.Close()
	bs, _ := ioutil.ReadAll(file)
	json.Unmarshal(bs, &DBInfo)
	fmt.Println("db info:", DBInfo)
}

func LoadFolderInfo() {
	path := "./DBConfigFile"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Could not open ", path, err)
		return
	}
	defer file.Close()

	bs, _ := ioutil.ReadAll(file)
	json.Unmarshal(bs, &lport)
	json.Unmarshal(bs, &FolderDir)
	fmt.Println("db info:", FolderDir)
}

/*
	Return the string we need to connect to db
	string we need in sql.open as second argument
	example :: sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/onlineEditor")
*/
func DBLogInString() string {
	return DBInfo.DB.Username + ":" + DBInfo.DB.Password + "@/" + DBInfo.DB.DBName
}
