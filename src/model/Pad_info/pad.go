// pad.go

package Pad

import (
	"io/ioutil"
	"errors"
	"fmt"
	"os"
)

type Pad_update struct{
	Value string 	`json:"value"`
	Start	uint 	`json:"start"`
	End	uint 	`json:"end"`
}

type Pad_info struct{
	ID string 	`json:"id"`
	Name string 	`json:"name"`
	Value string 	`json:"value"`
	Updates []Pad_update
	Needs_flushing bool
}

func (p *Pad_info) Add_update(v string , s , e uint ){
	p.Updates = append(p.Updates , Pad_update{v, s ,e })
}

func (p *Pad_info) Rmv_Updates(){
	p.Updates = nil
}


/*
	Get from file to pad.Value contents of file
*/
func (p *Pad_info) Get_Contents()(er error){
	
	filePath := "./SavedFiles/"+p.ID+".txt"
	file,er:= os.Open(filePath)
	if er!=nil{
		fmt.Println("Error opening ",filePath, "\n", er)
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

/*	Return a sub-string of value of pad according to start & end offsets	*/
func (p Pad_info) Get_Part(start , end int) (s string , er error){
	if (start < 0 || start > len(p.Value) || end < 0 || end > len(p.Value) ){
		er = errors.New("Out of bounds")
	}
	s = p.Value[start:end]

	return 
}


/*
	Try and write in file contetns of pad
	(update in filesystem)
	return er!=nil if failed
*/
func (p *Pad_info) Update_file() (er error){
	if p.Needs_flushing{
		filePath := "./SavedFiles/"+p.ID+".txt"
		if er = ioutil.WriteFile(filePath, []byte(p.Value), 0666); er!=nil{
			fmt.Println("Could not update file ", p.ID ,er)
		}else{
			p.Needs_flushing = false
		}

	}

	return
}
