// model.go

package Requests


// struct to parse PUT request json
type put_req string

const (
	Dlt put_req = "dlt"
	Wr  put_req = "wr"
)

// struct for decoding PUT request json from client   
type Client_Put struct {
	// for out-of-order requests
	Timestamp int 			`json:"Req_date"`

	// value for witing/inserting
	Val string 				`json:"Value"`

	// offset for write/start-of-insert/start-of-delete
	OffsetFrom uint 		`json:"Start"`

	// offset for end-of-insert/end-of-delete
	OffsetTo uint 			`json:"End"`

	// notepadID request is referring to
	Notepad_ID string		`json:"Pad_ID"`

	// signal for polling if true send back Pad updates
	Is_update_request bool	`json:"is_update"`
}


// struct for using client JSON for server use 
type Editor_req struct {
	// for out-of-order requests
	Timestamp int 			

	// value to be inserted
	Val string

	// offset for start-of-insert/delete
	OffsetFrom uint

	// offset for end-of-insert/delete
	OffsetTo uint

	// notepad ID request is referring to
	Notepad_ID string

	// user IP address
	UserIp	string
}

type Oldest_First []Editor_req

func (reqs Oldest_First) Len() int { return len(reqs) }
func (reqs Oldest_First) Swap(i,j int) { reqs[i] , reqs[j] = reqs[j] ,reqs[i] }
func (reqs Oldest_First) Less(i,j int) bool {
	return reqs[i].Timestamp < reqs[j].Timestamp
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