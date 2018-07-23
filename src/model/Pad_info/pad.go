// pad.go

package Pad

import (
	"fmt"
	"os"
	"io/ioutil"
)

//----------------------Options for Pad---NewPad(StorePad),Delete,Rename,EmptyDocument---------------------------
type Pad_info struct{
	ID string 	`json:"id"`
	Name string 	`json:"name"`
	Value string 	`json:"value"`
}


/*
	Get from file to pad.Value contents of file
*/
func (p *Pad_info) Get_Contents()(er error){
	
	filePath := "./SavedFiles/"+p.ID+".txt"
	file,er:= os.Open(filePath)
	if er!=nil{
		fmt.Println("Error opening ",filePath," check if exists")
		return
	}

	data , er := ioutil.ReadAll(file)
	if er!= nil{
		fmt.Println("Error reading contents of ",filePath)
		return
	}

	p.Value = string(data)
	return
}


/*
	Try and write in file contetns of pad
	(update in filesystem)
	return er!=nil if failed
*/
func (p Pad_info) Update_File() (er error){
	filePath := "./SavedFiles/"+p.ID+".txt"
	er = ioutil.WriteFile(filePath, []byte(p.Value), 0666)
	return
}