package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
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
	PadMap[t.ID]=&Pad{
		t.ID,
		t.Name,

	}

	fmt.Println(PadMap["zxc23"].ID)

}



func main(){

	router := httprouter.New()
	router.POST("/",store_pad)

	log.Fatal(http.ListenAndServe(":8080" ,router))

}

