package optionss

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	"os"
       _ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/lucasjones/reggen"
	"strconv"

)
type Pad struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Value string `json:"value"`
}


type Optionss_Fun interface {
	StorePad(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	RenameFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	DeleteFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)
	EmptyDocument(w http.ResponseWriter ,r *http.Request, _ httprouter.Params)

}
var (
	fileInfo *os.FileInfo
	err      error
)


func NewPad() *Pad{
	return  &Pad{}
}

type PadRequest struct{
        Id string `json:"id"`
}



var i=0

var PadMap=make(map[string]*Pad)


func (p Pad) StorePad (w http.ResponseWriter ,r *http.Request, _ httprouter.Params){

	fmt.Fprint(w,"Test1\n")
   db, err := sql.Open("mysql",
                "root:root@tcp(127.0.0.1:3306)/onlineEditor")
        if err != nil {
                panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
        }
        defer db.Close()


	s:=strconv.Itoa(i)
	s="Newpad"+s

	for{
		str, err2 := reggen.Generate("[a-f0-9]{16}", 16)
		if err2 != nil {
			panic(err2)
		}
		fmt.Println("AAAA",str)
		if val,ok :=PadMap[str]; ok {
			fmt.Println("Found",val.Name)

		}else{
			PadMap[str]=&Pad{
				str,
				s,
				"",
			}
			f:="./SavedFiles/"+str+".txt"
			os.Create(f)

			stmt,err := db.Prepare("INSERT INTO filesMetaData SET id=? , name=?")
			if err != nil {
				panic(err)

		}
			_, err = stmt.Exec(str, s)
			if err != nil {
					panic(err)
			}

			i=i+1
			break;
			}


   }






	for k, v := range PadMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}

	fmt.Printf("----------\n")

}


func (p Pad) RenameFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	fmt.Fprint(w,"Test1\n")
db, err := sql.Open("mysql",
                "root:root@tcp(localhost:3306)/onlineEditor")
        if err != nil {
                panic(err.Error())  // Just for example purpose. You should use proper error handling instead of panic
        }
        defer db.Close()


	decoder := json.NewDecoder(r.Body)
	var t Pad
	err = decoder.Decode(&t)
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
 _,err = db.Exec("UPDATE set name=$1 WHERE id=$2 ",t.Name,t.ID)
   if err != nil {
       panic(err)
   }

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func (p Pad) DeleteFile(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var t Pad
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
			fmt.Println("Delete", val.Name)
		err := os.Remove("./SavedFiles/"+PadMap[t.ID].ID+".txt")
		if err != nil {
			log.Fatal(err)
		}
		delete(PadMap,t.ID)
		

	}else{
		fmt.Println("File %s not found",t.ID)
	}
}

func (p Pad) EmptyDocument(w http.ResponseWriter ,r *http.Request, _ httprouter.Params){
	decoder := json.NewDecoder(r.Body)
	var t Pad
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}

	if val,ok :=PadMap[t.ID]; ok {
		fmt.Println("Empty Document : ", val.Name)
		err := os.Truncate("./SavedFiles/"+PadMap[t.ID].ID+".txt", 0)
		if err != nil {
			log.Fatal(err)
		}
	}else{
		fmt.Println("File %s not found",t.ID)
	}

}



