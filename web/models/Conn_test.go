package models

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"testing"
)

func TestConn(t *testing.T) {
	db, err := Conn("10.227.18.247","agent")
	if err != nil {
		fmt.Printf("%v",err)
	}
	d := []bson.M{}
	err = db.C("agent").Find(nil).All(&d)
	if err != nil {
		fmt.Printf("%v",err)
	}
	fmt.Print(d)
}
