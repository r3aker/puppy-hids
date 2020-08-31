package main

import "gopkg.in/mgo.v2/bson"

type MetaData struct {
	Name        string `json:"name" bson:"name"`
	Author      string `json:"author" bson:"author"`
	Description string `json:"description" bson:"description"`
	Level       int    `json:"level" bson:"level"`
}
type RulesData map[string]rule

func insert2rules() {
	metadata := MetaData{Name: "可疑文件操作" , Author: "thonsun", Description:"关键文件位置存在可疑文件操作" , Level: 0}
	ruledata := RulesData{
		"action": {
			Type: "regex",
			Data: "write|create|remove|chmod|chgrp",
		},
		// TODO：可以更细致一些 如chmod +s /etc/resolv.conf dns 解析的误报消除
		"path": {
			Type: "non-regex",
			Data: "/etc/resolv.conf.*",// 白名单
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"file","system":"linux","and":true,"enabled":true})

	// 单独的docker 文件变动检测
	metadata = MetaData{Name: "docker可疑文件操作" , Author: "thonsun", Description:"docker关键文件位置存在可疑文件操作" , Level: 0}
	ruledata = RulesData{
		"action":{
			Type:"regex",
			Data:"write|create|remove|chmod|chgrp",
		},
		"path":{
			Type:"non-regex",
			///var/lib/docker/overlay2/f553c1fceb7ba14cc8cda6e7f4aba27493c11f3c34cfa05d44ce5c70d97233d8/merged/sbin/thonsun
			Data:"/var/lib/docker/overlay2/\\w*?/merged/(etc/resolv.conf.*)",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"file","system":"linux","and":true,"enabled":true})

	metadata = MetaData{Name: "超级系统用户变化" , Author: "thonsun", Description:"存在非root的gid=0的用户，可能为攻击者创建" , Level: 0}
	ruledata = RulesData{
		"name": {
			Type: "non-regex",
			Data: "^root$",
		},
		"gid": {
			Type: "string",
			Data:"0",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"userlist","system":"linux","and":true,"enabled":true})

	metadata = MetaData{Name: "超级系统用户变化" , Author: "thonsun", Description:"存在非root的uid=0的用户，可能为攻击者创建" , Level: 0}
	ruledata = RulesData{
		"name": {
			Type: "non-regex",
			Data: "^root$",
		},
		"uid": {
			Type: "string",
			Data:"0",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"userlist","system":"linux","and":true,"enabled":true})


	metadata = MetaData{Name: "php webshell 行为" , Author: "thonsun", Description:"php www执行shell,可能是webshell文件" , Level: 0}
	ruledata = RulesData{
		"uid": {
			Type: "string",
			Data: "33",
		},
		"name": {
			Type: "regex",
			Data:"/bin/bash|/bin/sh|/bin/dash|/usr/bin/env",
		},
		"path":{
			Type:"regex",
			Data:"/var/www/*",
		},
		// docker 中进程的path 与 ppath 采集不出来
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"process","system":"linux","and":true,"enabled":true})

	// cat | touch | echo  xxx 黑名单细致
	metadata = MetaData{Name: "docker php webshell 行为" , Author: "thonsun", Description:"docker php www执行shell,可能是webshell文件" , Level: 0}
	ruledata = RulesData{
		"uid": {
			Type: "string",
			Data: "33",
		},
		"name": {
			Type: "regex",
			Data:"/bin/bash|/bin/sh|/bin/dash|/usr/bin/env",
		},
		// docker 中进程的path 与 ppath 采集不出来
		"path":{
			Type:"string",
			Data:"",
		},
		"ppath":{
			Type:"string",
			Data:"",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"process","system":"linux","and":true,"enabled":true})

	//jenkins shell 执行
	metadata = MetaData{
		Name:        "Jenkins bash命令执行",
		Author:      "thonsun",
		Description: "Jenkins bash命令执行，可能为RCE漏洞利用",
		Level:       0,
	}
	ruledata = RulesData{
		"pcmdline":{
			Type:"regex",
			Data:".*?/usr/share/jenkins/jenkins.war.*",
		},
		"cmdline":{
			Type:"regex",
			Data:"ls|touch|cat|echo|curl|wget|whoami|pwd|id",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"process","system":"linux","and":true,"enabled":true})

	metadata = MetaData{Name:"主机可疑外网连接", Author:"thonsun", Description: "主机主动发起外网连接", Level:0}
	ruledata = RulesData{
		"dir": {
			Type:"string",
			Data:"out",
		},
		"remote": {
			Type:"non-regex",
			Data:"(127\\.0\\.0\\.1)|(localhost)|(10\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})|(172\\.((1[6-9])|(2\\d)|(3[01]))\\.\\d{1,3}\\.\\d{1,3})|(192\\.168\\.\\d{1,3}\\.\\d{1,3})",
		},
		"cmdline":{
			Type:"non-regex",
			Data:"^/usr/.*|^/var/lib.*|^/bin/.*|^/sbin/.*",
		},
		"port":{
			Type:"non-regex",
			Data:"^80$|^443$|^8080$",
		},
	}
	DB.C("rules").Insert(bson.M{"meta":metadata,"rules":ruledata,"source":"connection","system":"all","and":true,"enabled":true})

}
