package Users

import (
	"time"
)

type User struct{
	Address string
	LastActive time.Time
}

func (u *User) KeepActive(){	
	u.LastActive = time.Now();
}

/*	
	Checks if user is active 
	A user is considered to be inactive if he has not   
	editted anything for the last 3 mins
*/
func (u *User) IsActive() bool{
	// expiration occurs 3 minutes after the time user was last active
	expirationLimit := u.LastActive.Add(time.Minute*time.Duration(3));
	
	// if expiration time has passed return user is inactive
	if expirationLimit.Before(time.Now()){
		return false
	}

	// user is active
	return true
}


/*	timeout implementation
	Checks ConnectedUsers:
		If for any user the inactive period has
		exceed allowed remove him form
		ConnectedUsers slice
*/
func CleanInactiveUsers(){
	for padID , users := range ConnectedUsers{
		for idx , u := range users{
			if !u.IsActive(){
				ConnectedUsers[padID] = append(ConnectedUsers[padID][:idx], 
					ConnectedUsers[padID][idx+1:]...)
			}
		}
	}	
}

/*
	A global map to keep all the users that editing
	a pad.
	Key is a pad id(string) and as value 
	we have a slice that we keep all connected users
*/
var ConnectedUsers = make(map[string][]User)

/*insert a new userIp according to padId, to the map */
func InsertUserIp(ip string, padId string) {
	ConnectedUsers[padId] = append(ConnectedUsers[padId], 
		User{ip , time.Now() } )
}

/*Delete a userIp according to padId, from the map */
func DeleteUserIp(ip string, padId string) {
	for i := 0; i < len(ConnectedUsers[padId]); i++ {
		u := ConnectedUsers[padId][i]
		if u.Address == ip {
			ConnectedUsers[padId] = append(ConnectedUsers[padId][:i], ConnectedUsers[padId][i+1:]...)
			break
		}
	}
}

/*
	Return a slice with all the users' ip that
	are editting the pad. 
	If no users are editting the pad then return an
	empry slice
*/
func GetConnectedUsers(padId string) []User {
	return ConnectedUsers[padId]
}
