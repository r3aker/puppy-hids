package action

import (
	"github.com/thonsun/puppy-hids/server/models"
	"github.com/thonsun/puppy-hids/server/utils"
	"gopkg.in/mgo.v2/bson"
)
type client struct {//mongodb data
	TYPE string
	DIC ClientConfig
}
// ClientConfig 客户端配置信息结构
type ClientConfig struct {
	Cycle       int      `bson:"cycle"` // 信息传输频率，单位：分钟
	UDP         bool     `bson:"udp"`   // 是否记录UDP请求
	LAN         bool     `bson:"lan"`   // 是否本地网络请求
	Type        string   `bson:"type"`  // 模式，考虑中
	Filter      filter   // 直接过滤不回传的数据
	MonitorPath []string `bson:"monitorPath"` // 监控目录列表
	Lasttime    string   // 最后一条登录日志时间
}

type filterres struct {//mongodb data
	Type string `bson:"type"`
	Dic  filter `bson:"dic"`
}
type filter struct {
	File    []string `bson:"file"`    // 文件hash、文件名
	IP      []string `bson:"ip"`      // IP地址
	Process []string `bson:"process"` // 进程名、参数
	Port 	[]int	`bson:"port"` // 端口白名单
}

//从mongodb 获取配置
func GetAgentConfig(ip string) (config ClientConfig) {
	//test
	//config = ClientConfig{
	//	Cycle:       0,
	//	UDP:         false,
	//	LAN:         false,
	//	Type:        "",
	//	Filter:      filter{},
	//	MonitorPath: []string{"/home/thonsun/workspace/puppy/*"},
	//	Lasttime:    fmt.Sprintf("%d", time.Now()),
	//}

	var clientRes client
	c := models.DB.C("config")
	//TODO: 细化到IP 具体机器的配置
	c.Find(bson.M{"type":"client"}).One(&clientRes)// 全局唯一的配置，linux | windows 都在这里面
	config = clientRes.DIC

	var res filterres
	c.Find(bson.M{"type": "filter"}).One(&res)
	config.Filter = res.Dic

	// TODO：获取最后一条登录日志
	// 分多次实现嵌套struct的初始化
	//lastTime, err := models.QueryLogLastTime(ip)
	//if err != nil {
	//	utils.Error("query %v last login log error:%v",ip,err.Error())
	//	config.Lasttime = "all"
	//} else {
	//	config.Lasttime = lastTime
	//}
	config.Lasttime = "all"
	utils.Debug("获取agent %v 配置: %#v",ip,config)

	return config
}
