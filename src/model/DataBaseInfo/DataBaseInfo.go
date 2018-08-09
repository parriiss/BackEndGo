package DataBaseInfo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//a struct to store the db info for log in
type DataBaseInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
	Ip       string `json:"ip"`
}

// a global instance of DataBaseInfo
var DBInfo DataBaseInfo

/*
	function to read from DBconfigFile 
	json file the info of db
*/
func LoadDBInfo() {
	path := "./DBConfigFile"
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Could not open ",path , err)
		return
	}
	defer file.Close()
	
	bs, _ := ioutil.ReadAll(file)
	json.Unmarshal(bs, &DBInfo)
	fmt.Println("db info:" ,DBInfo)
}

/*
	Return the string we need to connect to db
	string we need in sql.open as second argument
	example :: sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/onlineEditor")
*/
func DBLogInString() string {
	return DBInfo.Username+ ":"+DBInfo.Password+"@/"+DBInfo.DBName
}
