package model

import (
        "fmt"
        "encoding/json"
        "net/http"
        "github.com/julienschmidt/httprouter"
)






//--------PARIS-------New Pad----
type Pad struct{
ID string json:'id'
Name string json:'name'
Last_edited string json:'last_edited'
Size int json:'size'
}
var PadMap=make(map[string]*Pad)


func store_pad(w http.ResponseWriter ,r *http.Request, p httprouter.Params){
PadMap[id]=&Pad{
id,
Name.
Last_edited,
Size,
}


}



//-----------PARIS_END------------


