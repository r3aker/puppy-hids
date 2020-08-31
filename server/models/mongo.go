package models

import (
	"github.com/thonsun/puppy-hids/server/utils"
	"gopkg.in/mgo.v2"
)

func conn(netloc string, dbname string) (*mgo.Database, error) {
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

func mgoCheck() {
	utils.Debug("check mongodb...")
	err := DB.Session.Ping()
	if err != nil {
		utils.Debug(err.Error())
		DB.Session.Refresh()
	}
}
