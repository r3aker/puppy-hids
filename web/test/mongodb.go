package main

import "gopkg.in/mgo.v2"

func conn(dbname string) (*mgo.Database, error) {
	info := mgo.DialInfo{
		Addrs:          []string{mongohome},
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