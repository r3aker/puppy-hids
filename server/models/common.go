package models

import (
	"puppy-hids/server/utils"
	"flag"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"os"
	"strings"
	"time"
)

var (
	DB *mgo.Database
	mongodb *string
	es *string

	// server的配置：如要告警｜检测配置
	Config serverConfig

	LocalIP string
	err error
	// 存放在mongodb rule的规则库
	RuleDB = []ruleInfo{}
)

type serverConfigres struct {
	Type string       `bson:"type"`
	Dic  serverConfig `bson:"dic"`
}
type serverConfig struct {
	Learn bool `bson:"learn"`
	OfflineCheck bool `bson:"offlinecheck"`
	BlackList blackList
	WhiteList whiteList

	Private string `bson:"privatekey"`
	Cert string `bson:"cert"`
}

// 从 mongo中取出配置
type whiteListres struct {
	Type string `bson:"type"`
	Dic whiteList `bson:"dic"`
}
type blackListres struct {
	Type string `bson:"type"`
	Dic blackList `bson:"dic"`
}
type whiteList struct {
	File []string `bson:"file"`
	IP []string `bson:"ip"`
	Process []string `bson:"process"`
	Other []string `bson:"other"`
}
type blackList struct {
	File []string `bson:"file"`
	IP []string `bson:"ip"`
	Process []string `bson:"process"`
	Other []string `bson:"other"`
}

/*
   {
       "and": true,
       "enabled": true,
       "meta": {
           "author": "wolf",
           "description": "带有$符号的隐藏用户,很大可能为黑客设置的隐藏后门账户",
           "level": 0,
           "name": "隐藏用户"
       },
       "rules": {
           "name": {
               "data": "\\$",
               "type": "regex"
           }
       },
       "source": "userlist",
       "system": "windows"
   },
*/
type rule struct {
	Type string `json:"type" bson:"type"` // 是否要regex 配置
	Data string `json:"data" bson:"data"`
}

type ruleInfo struct {
	Meta struct {
		Name        string `json:"name" bson:"name"`               // 名称
		Author      string `json:"author" bson:"author"`           // 编写人
		Description string `json:"description" bson:"description"` // 描述
		Level       int    `json:"level" bson:"level"`             // 风险等级
	} `json:"meta" bson:"meta"` // 规则信息
	Source string          `json:"source" bson:"source"` // 选择判断来源
	System string          `json:"system" bson:"system"` // 匹配系统
	Rules  map[string]rule `json:"rules" bson:"rules"`   // 具体匹配规则
	And    bool            `json:"and" bson:"and"`       // 规则逻辑
}

func init() {
	mongodb = flag.String("db", "", "mongodb ip:port")
	es = flag.String("es", "", "elasticsearch ip:port")
	flag.Parse() // 绑定解析变量
	if len(os.Args) <= 2 {
		flag.PrintDefaults() // 输出help 信息
		os.Exit(1)
	}
	if strings.HasPrefix(*mongodb, "127.") || strings.HasPrefix(*mongodb, "localhost") {
		utils.Error("mongodb Can not be 127.0.0.1")
		os.Exit(1)
	}
	DB, err = conn(*mongodb, "agent") // 只有agent这个数据库
	if err != nil {
		utils.Error("init mongo client error:%v",err)
		flag.PrintDefaults()
		os.Exit(1)
	}
	LocalIP, err = getLocalIP(*mongodb)
	if err != nil {
		utils.Error("set local ip error:",err)
		os.Exit(1)
	}
	utils.Debug("Get Config: serverconfig monitor rules etc")
	setConfig()
	setRules()
	go esCheckThread() //一小时检查一次 ：当前写入索引是否要换了
}

// setConfig 获取配置文件
func setConfig() {
	utils.Debug("get server config...")
	c := DB.C("config")
	res := serverConfigres{}
	c.Find(bson.M{"type": "server"}).One(&res)

	res3 := blackListres{}
	c.Find(bson.M{"type": "blacklist"}).One(&res3)

	res4 := whiteListres{}
	c.Find(bson.M{"type": "whitelist"}).One(&res4)


	Config = res.Dic
	// cert | private 等非嵌入json已经在这里初始化了
	Config.BlackList = res3.Dic
	Config.WhiteList = res4.Dic
}

// setRules 获取异常规则集
func setRules() {
	utils.Debug("get monitor rules...")
	c := DB.C("rules")
	c.Find(bson.M{"enabled": true}).All(&RuleDB) // 批量的获取是bson的一个数组,enabled 是选择rule,由后端写入，不存在ruleinfo结构体中

}

// regServer 注册为服务，Agent才知道发给谁
func regServer() {
	utils.Debug("register server...")
	c := DB.C("server")
	_, err := c.Upsert(bson.M{"netloc": LocalIP + ":33433"}, bson.M{"$set": bson.M{"uptime": time.Now()}})
	if err != nil {
		utils.Debug("register to mongo error:%v",err.Error())
	}
}

func getLocalIP(ip string) (string, error) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

// Heartbeat server心跳线程,定期写入server 存活，定时刷新配置和规则
func Heartbeat() {
	for {
		// TODO:并发线程更新配置 安全
		utils.Debug("Start heartbeat thread...")
		mgoCheck()  // mongodb 检测
		regServer() // 向mongodb注册server IP 端口
		setConfig() // 从mongodb获取配置,定时更新
		setRules()
		time.Sleep(time.Second * 30)
	}
}

