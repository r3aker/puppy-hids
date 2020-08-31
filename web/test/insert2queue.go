package main

import (
	"time"
)

var home string = "192.168.8.149"
var home1 string = "192.168.8.114"
// 保存到mongodb 中的数据格式
type queue struct {
	IP string `bson:"ip"`
	Type string `bson:"type"`
	Command string `bson:"command"`
	Time time.Time `bson:"time"`
}

func insert2queue(){
	//q := queue{
	//	IP:      devbox,
	//	Type:    "exec",
	//	Command: "ifconfig",
	//	Time:    time.Now(),
	//}

	//q := queue{
	//	IP:      devbox,
	//	Type:    "update",
	//	Command: "",
	//	Time:    time.Now(),
	//}

	//q := queue{
	//	IP:      devbox,
	//	Type:    "kill",
	//	Command: "vim",
	//	Time:    time.Now(),
	//}

	//q := queue{
	//	IP:      home,
	//	Type:    "reload",
	//	Command: "",
	//	Time:    time.Now(),
	//}

	q := queue{
		IP:      home,
		Type:    "uninstall",
		Command: "",
		Time:    time.Now(),
	}

	//q := queue{
	//	IP:      devbox,
	//	Type:    "stop",
	//	Command: "",
	//	Time:    time.Now(),
	//}

	//q := queue{
	//	IP:      devbox,
	//	Type:    "continue",
	//	Command: "",
	//	Time:    time.Now(),
	//}
	DB.C("queue").Insert(q)
}