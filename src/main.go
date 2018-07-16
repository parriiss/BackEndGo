// main.go

package main

/* 	
	import folders that have func implementetions 
	for requests
 */

import(
	_ "fmt"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"./Controller"
)


func main() {
	
	r := httprouter.New()

// init API
	handleURLS(r)

// fire up server
	http.ListenAndServe(":8000" , r)
}




// Assign functions to handle requests
func handleURLS(r *httprouter.Router){
	c := controll.NewController()

// 	GET
	r.GET("/OnlineEditor/About" , c.About)
	


// 	POST
// r.POST(<URL1> , <function>)
	// ....
	// ...
	// . 


// 	PUT
// r.PUT(<URL1> , <function>)
	// ....
	// ...
	// . 


//	DELETE
// r.DELETE(<URL1> , <function>)
	// ....
	// ...
	// . 
}