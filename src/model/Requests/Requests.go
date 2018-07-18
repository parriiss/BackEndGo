// model.go

package Requests

import "time"
// MOVE THESE  STRUCTS TO MODEL
// struct to parse PUT request json
type put_req string

const (
	Ins put_req = "ins"
	Dlt put_req = "dlt"
	Wr  put_req = "wr"
)

type Client_Put struct {
	// for out-of-order requests
	Req_date time.Time 	`json:"Req_date"`

	// value for witing/inserting
	Val string 		`json:"Value"`

	// offset for write/start-of-insert/start-of-delete
	OffsetFrom int 		`json:"Start"`

	// offset for end-of-insert/end-of-delete
	OffsetTo int 		`json:"End"`
}

type Client_Dlt struct {
	// for out-of-order requests
	Req_date time.Time 	`json:"Req_date"`

	// offset for write/start-of-insert/start-of-delete
	OffsetFrom int 		`json:"Start"`

	// offset for end-of-insert/end-of-delete
	OffsetTo int 		`json:"End"`
}

// struct for decoding JSON from client
type Editor_req struct {
	// for out-of-order requests
	Req_date time.Time

	// what type of update server is doing
	// unecessary??
	Req_type put_req

	// value for writing/inserting
	Val string

	// offset for write/start-of-insert/start-of-delete
	// negative vals??
	OffsetFrom int

	// offset for end-of-insert/end-of-delete
	// negative vals??
	OffsetTo int
}

// channel for parsing req to req handling routine
var In chan Editor_req = make(chan Editor_req)

// open channel for request parsing
func Init(){
	In = make(chan Editor_req)	
}

// close request parsing channel 
func CloseChannel(){
	close(In)
}