package PadHistory

//a struct to store padHistory info from db
type PadHistory struct {
	Ip    string `json:"ip"`
	State int    `json:"state"`
	Time  string `json:"time"`
}
