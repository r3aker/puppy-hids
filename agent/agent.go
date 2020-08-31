package main

import (
	"github.com/thonsun/puppy-hids/agent/client"
	"github.com/thonsun/puppy-hids/agent/common/log"
	"github.com/thonsun/puppy-hids/agent/config"
	"fmt"
	"os"
)

func CreateLogFile() {
	var err error
	var logfile *os.File
	err = os.MkdirAll(config.LOGFILE_PATH,os.ModePerm)
	if err != nil {
		log.Error("create log path error:%v",err)
		return
	}
	//TODO:根据日期创建日志文件
	logfile, err = os.OpenFile(config.LOGFILE,os.O_APPEND | os.O_CREATE |os.O_RDWR,0666)
	if err != nil {
		log.Error("create log file error:%v",err)
		return
	}
	log.SetOutput(logfile)
}
func main() {
	CreateLogFile()
	if len(os.Args) <= 1 {
		fmt.Println("Usage: agent web IP [debug]")
		fmt.Println("Example: agent 192.168.0.1:8000 debug")
		return
	}
	var agent client.Agent
	agent.ServerNetLoc = os.Args[1]
	if len(os.Args) == 3 && os.Args[2] == "debug" {
		log.SetLogLevel(log.DEBUG)
		log.Debug("DEBUG MODE")
	}else {
		log.SetLogLevel(log.INFO)
	}
	agent.Run()
}
