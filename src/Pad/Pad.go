package Pad

import (
	//"fmt"
)

type PadRequest struct{
	Id string `json:"id"`
}


type Pad struct{
	Id string `json:"id"`
	Name string `json:"name"`
	Value string `json:"value"`
}
var PadMap=make(map[string]*Pad)



/*
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
*/

/*
func main(){

	router := httprouter.New()
	router.POST("/",store_pad)

	log.Fatal(http.ListenAndServe(":8080" ,router))

}
*/
