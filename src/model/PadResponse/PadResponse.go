package PadResponse

import "../Users"

type ClientR struct{
	// ID of the pad
	ID string 		`json:"id"`

	// name of the pad
	Name string 	`json:"name"`

	// Contents
	Value string 	`json:"value"`

	//  Users Connected to this pad
	Users []Users.User	`json:"users"`
} 