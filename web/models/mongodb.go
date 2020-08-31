package models

import (
	"gopkg.in/mgo.v2"
)

func Conn(netloc string, dbname string) (*mgo.Database, error) {
	info := mgo.DialInfo{
		Addrs:          []string{netloc},
		Username:       "root",
		Password:       "root",
	}
	session, err := mgo.DialWithInfo(&info)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	db := session.DB(dbname)
	return db, nil
}
