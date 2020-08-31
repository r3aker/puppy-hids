package main

import "gopkg.in/mgo.v2"

const (
	mongohome string = "192.168.8.114:27017"
	db = "agent"
)

var DB *mgo.Database
var err error

