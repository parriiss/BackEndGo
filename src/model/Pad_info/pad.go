// pad.go

package Pad

import (
	"io/ioutil"
	"fmt"
	"os"
)

//----------------------Options for Pad---NewPad(StorePad),Delete,Rename,EmptyDocument---------------------------
type Pad_info struct{
	ID string 	`json:"id"`
	Name string 	`json:"name"`
	Value string 	`json:"value"`
	Need_upd bool 
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
func (p *Pad_info) Update_file() (er error){
	if p.Need_upd{
		filePath := "./SavedFiles/"+p.ID+".txt"
		if er = ioutil.WriteFile(filePath, []byte(p.Value), 0666); er!=nil{
			fmt.Println("Could not update file ", p.ID ,er)
		}else{
			p.Need_upd = false
		}
	}
	return
}

func (p *Pad_info) Need_update(){
	p.Need_upd = true
}