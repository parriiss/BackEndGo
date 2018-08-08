package LogedInUsers

/*
	A global map to keep all the users that editing
	a same pad. Key is a pad id(string) and as value 
	we have a slice that we keep all ip(string)
	of users editing the pad
*/
var LogedInUsers = make(map[string][]string)

/*insert a new userIp according to padId, to the map */
func InsertUserIp(ip string, padId string) {
	LogedInUsers[padId] = append(LogedInUsers[padId], ip)
}

/*Delete a userIp according to padId, from the map */
func DeleteUserIp(ip string, padId string) {
	for i := 0; i < len(LogedInUsers[padId]); i++ {
		if LogedInUsers[padId][i] == ip {
			LogedInUsers[padId] = append(LogedInUsers[padId][:i], LogedInUsers[padId][i+1:]...)
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
func GetUsers(padId string) []string {
	return LogedInUsers[padId]
}
