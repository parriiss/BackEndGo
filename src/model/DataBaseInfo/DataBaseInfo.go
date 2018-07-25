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
	Port     string `json:"port"`
	DBName   string `json:"dbName"`
	Ip       string `json:"ip"`
}

// a global instance of DataBaseInfo
var DBInfo DataBaseInfo

/*
	function to read from DBconfigFile
	the info of db
	the infos in DBconfigFile are in json format
*/
func LoadDBInfo() {
	file, err := os.Open("ConfigFiles/DBConfigFile")
	if err != nil {
		fmt.Println("error in DBConfigFile")
		return
	}
	defer file.Close()
	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &DBInfo)
}
