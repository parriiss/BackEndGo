package model

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


type PadRequest struct{
        Id string `json:"id"`
}




type Pad struct{
	ID string `json:"id"`
	Name string `json:"name"`
}
var PadMap=make(map[string]*Pad)


func store_pad(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"Test1\n")
   db, err := sql.Open("mysql",
                "root:root@tcp(localhost:3306)/test1")
        if err != nil {
                panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
        }
        defer db.Close()


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

 _,err = db.Exec("INSERT INTO filesMetaData (id,name) VALUES (%s,%s)",t.ID,t.Name)
   if err != nil {
       panic(err)
   }





	}

	for k, v := range PadMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}
fmt.Printf("----------\n")


}


func rename_file(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"Test1\n")
db, err := sql.Open("mysql",
                "root:root@tcp(localhost:3306)/test1")
        if err != nil {
                panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
        }
        defer db.Close()


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
 _,err = db.Exec("UPDATE set name=%s WHERE id=%s ",t.Name,t.ID)
   if err != nil {
       panic(err)
   }

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



