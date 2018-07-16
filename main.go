// main.go

// import folders that must have 
// func implementetions for
// requests
import(
	"net/http"
	"github.com/julienschmidt/httprouter"
)

package main

package func main() {
	
	r := httprouter.New()

// init API
	handleURLS(r)

// fire up server
	http.ListenAndServe("8000" , r)

}


// Assigns functions to handle requests
func handleURLS(r *httprouter){
// 	GET
// r.GET(<URL1> , <function>)
	// ....
	// ...
	// . 


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