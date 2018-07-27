// pad.go

package Pad

import (
	"io/ioutil"
	"fmt"
	"os"
	"sync"
)

//----------------------Options for Pad---NewPad(StorePad),Delete,Rename,EmptyDocument---------------------------
type Pad_info struct{
	ID string 	`json:"id"`
	Name string 	`json:"name"`
	Value string 	`json:"value"`
	Needs_flushing bool 
	can_write_to_file sync.Mutex
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
		p.can_write_to_file.Lock()
		filePath := "./SavedFiles/"+p.ID+".txt"
		if er = ioutil.WriteFile(filePath, []byte(p.Value), 0666); er!=nil{
			fmt.Println("Could not update file ", p.ID ,er)
		}else{
			p.Needs_flushing = false
		}
		p.can_write_to_file.Unlock()
	}
	return
}

func (p *Pad_info) Need_update(){
	p.Needs_flushing = true
}