// pad.go

package Pad

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"../DataBaseInfo"
	"../Users"
)

type Pad_update struct {
	Value string `json:"value"`
	Start uint   `json:"start"`
	End   uint   `json:"end"`
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

// Append new update to slice in pad that keeps updates
// that have happened to later inform client at request
func (p *Pad_info) Add_update(v string, s, e uint) {
	p.Updates = append(p.Updates, Pad_update{v, s, e})
}

// Free the updates slice of pad
func (p *Pad_info) Rmv_Updates() {
	p.Updates = nil
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
			pad.Users = tmp
			PadMap[padID] = pad
		}else{
			// must add a 
			fmt.Println(time.Now() , "\nPad with Id:" , padID,"has no users connected to it, removed from mem")
			delete(PadMap , padID)
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
