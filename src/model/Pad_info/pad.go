// pad.go

package Pad

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"../DataBaseInfo"
	"../Users"
)

var PadLock sync.Mutex

type Pad_update struct {
	Value string `json:"value"`
	Start uint   `json:"start"`
	End   uint   `json:"end"`
	//  That must be notified
	ToNotify []Users.User
}


type Pad_info struct {
	// ID of the pad
	ID string `json:"id"`

	// name of the pad
	Name string `json:"name"`

	// Contents
	Value string `json:"value"`

	//  Updates that pad has gone through
	//  and users need to be notified about
	Updates []Pad_update

	//  Boolean pad is dirty must write to disk
	Needs_flushing bool

	//  Users Connected to this pad
	Users []Users.User
}

// Struct for sending pad to client 
type PadResponse struct{
	// ID of the pad
	ID string 		`json:"id"`

	// name of the pad
	Name string 	`json:"name"`

	// Contents
	Value string 	`json:"value"`

	//  Users Connected to this pad
	Users []Users.User	`json:"users"`
} 

// Append new update to slice in pad that keeps updates
// that have happened to later inform client at request
func (p *Pad_info) Add_update(v string, s, e uint ,toNotify []Users.User) {
	p.Updates = append(p.Updates, Pad_update{v, s, e, toNotify})
}

// Remove any updates that no user needs to be notified for
func (p *Pad_info) Rmv_Updates() {
	tmp := p.Updates[:0]
	for _ , u := range p.Updates{
		if (len(u.ToNotify)>0){
			tmp = append(tmp ,u)
		}else{
			fmt.Println("Notified all connected users about update:",u)
		}
	}

	PadLock.Lock()
	p.Updates = tmp
	PadLock.Unlock()
}

/*
	Get file contents to pad.Value
*/
func (p *Pad_info) Get_Contents() (er error) {

	filePath := "./SavedFiles/" + p.ID + ".txt"
	file, er := os.Open(filePath)
	if er != nil {
		fmt.Println("Error opening ", filePath, "\n", er)
		return
	}

	data, er := ioutil.ReadAll(file)
	if er != nil {
		fmt.Println("Error reading contents of ", filePath)
		return
	}

	p.Value = string(data)
	return
}

/*	Return a sub-string of value of pad according to start & end offsets	*/
func (p Pad_info) Get_Part(start, end int) (s string, er error) {
	if start < 0 || start > len(p.Value) || end < 0 || end > len(p.Value) {
		er = errors.New("Out of bounds")
	}
	s = p.Value[start:end]

	return
}

/*
	Try and write in file contetns of pad
	(update in filesystem)
	return er!=nil if failed
*/
func (p *Pad_info) Update_file() (er error) {
	if p.Needs_flushing {
		filePath := "./SavedFiles/" + p.ID + ".txt"
		if er = ioutil.WriteFile(filePath, []byte(p.Value), 0666); er != nil {
			fmt.Println("Could not update file ", p.ID, er)
		} else {
			p.Needs_flushing = false
		}

	}

	return
}

/*Exported map that holds the pads that are editted */
var PadMap = make(map[string]*Pad_info)

/*	timeout implementation 
	Checks ConnectedUsers:
		If for any user the inactive period has
		exceed allowed remove him form
		ConnectedUsers slice
*/
func CleanInactiveUsers() {
	for padID, pad := range PadMap {
		tmp := pad.Users[:0]
		for i, u := range pad.Users {
			// create a zero-length slice with the same underlying array

			if u.IsActive() {
				// keep element
				tmp = append(tmp, u)
			} else {
				if er := end_session(padID , u.Address); er!=nil{
					fmt.Println("Error signaling end session to db:\n\t" ,er)	
				}
				fmt.Println("Removing inactive user from pad:",
					padID, "\n\tUsers:", pad.Users, "\n\tlenght:", len(pad.Users)-(i-len(tmp)) )
				fmt.Println( /*"Removing idx:",idx,*/ "now:", time.Now(), "\nlastActive:", u.LastActive)
			}
		}
		
		// if there are active users for this pad
		if (len(tmp) > 0){
			// save to map the slice that keeps all active users
			PadLock.Lock()
			
			pad.Users = tmp
			PadMap[padID] = pad
			
			PadLock.Unlock()
		}else{
			fmt.Println(time.Now() , "\nPad with Id:" , padID,"has no users connected to it, removed from mem")
			PadLock.Lock()
			
			delete(PadMap , padID)
			
			PadLock.Unlock()
		}
	}
}

func end_session(id, ip string) (er error) {
	db, er := sql.Open("mysql", DataBaseInfo.DBLogInString())
	defer db.Close()
	
	stmt, err := db.Prepare("INSERT INTO historyFiles SET ip=?, id=?, time=?, state=?")
	if err != nil {
		return
	}
	_, err = stmt.Exec(ip, id, time.Now().Format("2006-01-02 15:04:05"), 0)
	if err != nil {
		return
	}

	return
}


/*Delete a userIp according to padId, from the map */
func DeleteUserIp(ip string, padId string) {
	var users []Users.User

	if _, ok := PadMap[padId]; !ok {
		return
	}

	users = PadMap[padId].Users
	for i := 0; i < len((users)); i++ {

		if users[i].Address == ip {
			// remove user
			users = append(users[:i], users[i+1:]...)

			//  update padmap check if pad has any other active users
			PadMap[padId].Users = users
			break
		}
	}
}

/*
	Return a slice with all the users' ip address
	that are editting the pad.
	If no users are editting the pad return nil
*/
func GetConnectedUsers(padId string) []Users.User {
	if _, ok := PadMap[padId]; ok {
		return PadMap[padId].Users
	}
	return nil
}

/*insert a new userIp according to padId, to the map */
func InsertUserIp(ip string, padId string) {
	PadMap[padId].Users = append(PadMap[padId].Users,
		Users.User{ip, time.Now()})
}

func GetUserUpdates(userAddress , padID string) (upd []Pad_update) {
	// test if need to update map with assign
	// probably not since map has pointers and not copy of values
	p := PadMap[padID]

	if len(p.Updates) == 0 {upd = nil; return}
	
	for uIDX := 0; uIDX<len(p.Updates); uIDX++{

		u := p.Updates[uIDX]
		for nIDX:=0; nIDX<len(u.ToNotify); nIDX++{
			if userAddress == u.ToNotify[nIDX].Address{
				upd = append(upd , u)
				/*
					user is notified
					 remove user from needtoNotify slice
				*/
				u.ToNotify = append(u.ToNotify[:uIDX] ,u.ToNotify[uIDX+1:]...)
				break
			}
		}
		
		// clean up update if there is noone left to notify after this 
		if len(u.ToNotify)==0{
			p.Updates = append(p.Updates[:uIDX] ,p.Updates[uIDX+1:]...)
			uIDX--	//check repeat index for slice after removing 1 item 
		}
	}
	
	return
}

/* 
	Get the users that need to be notified for update
	when user with ip address(uAddress) is making that 
	update 
*/
func GetUsersToNotify(uAddr, padId string) (notify []Users.User){
	p := PadMap[padId]
	for idx,user := range p.Users{
		if user.Address == uAddr{
			// append the rest after the person that is making upd
			notify = append(notify , p.Users[idx+1:]...)
			break;
		}
		notify = append(notify , user)
	}
	
	return
}
