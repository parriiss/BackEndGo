package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"os"
)

var (
	fileInfo *os.FileInfo
	err      error
)

type Pad struct{
	ID string `json:"id"`
	Name string `json:"name"`
}
var PadMap=make(map[string]*Pad)


func store_pad(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"Test1\n")

	decoder := json.NewDecoder(r.Body)
	var t Pad
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Found", val.Name)

	} else {
	PadMap[t.ID]=&Pad{
	t.ID,
	t.Name,

	}
	}

	for k, v := range PadMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}
fmt.Printf("----------\n")


}


func rename_file(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"Test1\n")

	decoder := json.NewDecoder(r.Body)
	var t Pad
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Found", val.Name)
		/*
		err := os.Rename(PadMap[t.ID].Name, t.Name)
		if err != nil {
			log.Fatal(err)
		}*/
		PadMap[t.ID].Name=t.Name

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func delete_file(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var t Pad
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
			fmt.Println("Delete", val.Name)
		delete(PadMap,t.ID)
		err := os.Remove(PadMap[t.ID].ID)
		if err != nil {
			log.Fatal(err)
		}

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}





func main(){

	router := httprouter.New()
	router.POST("/",store_pad)

	log.Fatal(http.ListenAndServe(":8080" ,router))

}

