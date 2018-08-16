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
