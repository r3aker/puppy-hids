package common

import (
	"puppy-hids/web/common/log"
	"puppy-hids/web/models"
	"crypto/md5"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"time"
)

const (
	MONGODB_HOME string = "192.168.8.114:27017"
)

var (
	DB *mgo.Database
	err error
)

type Config struct {
	Id bson.ObjectId `bson:"_id" json:"_id,omitempty"`
	Type string `bson:"type" json:"type"`
	Dic interface{} `bson:"dic" json:"dic"`
}

// 导出要大写
type Server struct {
	Netloc string `bson:"netloc"`
	Uptime  time.Time `bson:"uptime"`
}

func init() {
	//TODO: 环境变量设置mongodb
	mongodb_addr := os.Getenv("MONGODB")
	DB,err = models.Conn(mongodb_addr,"agent")
	if err != nil {
		log.Debug("connect mongodb error:%v",err)
	}
}

// 告警结构体
//{
//	"_id" : ObjectId("5ece8738f247488dae8a98f0"),
//	"type" : "process",
//	"source" : "webshell 行为",
//	"info" : "/bin/dash|33",
//	"description" : "www执行shell,可能是webshell文件",
//	"ip" : "10.227.18.247",
//	"level" : 0,
//	"status" : 0,
//	"raw" : "{\"auid\":\"4294967295\",\"cmdline\":\"sh\\u0000-c\\u0000id\\u0000\",\"comm\":\"sh\",\"epoch\":\"1590593336.269\",\"euid\":\"33\",\"exit\":\"0\",\"gid\":\"33\",\"key\":\"proc_create\",\"logid\":\"32307374\",\"name\":\"/bin/dash\",\"path\":\"\",\"pcmdline\":\"apache2\\u0000-D\\u0000FOREGROUND\\u0000\",\"pid\":\"754760\",\"pname\":\"apache2\\n\",\"ppath\":\"\",\"ppid\":\"751344\",\"success\":\"yes\",\"suid\":\"33\",\"syscall_id\":\"59\",\"uid\":\"33\"}",
//	"time" : ISODate("2020-05-27T15:28:56.275Z")
//}
type Notice struct {
	Type string `bson:"type"`
	Source string `bson:"source"`
	Description string `bson:"description"`
	IP string `bson:"ip"`
	Raw string `bson:"raw"`
	Info string `bson:"info"`
	Time time.Time `bson:"time"`
}

// 存活agent
//{
//"_id" : ObjectId("5ece3955f247488dae8910d5"),
//"ip" : "10.227.18.247",
//"hostname" : "n227-018-247",
//"path" : [ ],
//"system" : "Debian GNU/Linux 9 \\n \\l 4.14.81.bm.15-amd64\nx86_64\n",
//"type" : "",
//"uptime" : ISODate("2020-05-28T02:18:45.756Z"),
//"health" : 0
//}
type Client struct {
	IP string `bson:"ip"`
	Hostname string `bson:"hostname"`
	System string `bson:"system"`
	Uptime time.Time `bson:"uptime"`
}

// FileMD5String 获取文件MD5
func FileMD5String(filePath string) (MD5String string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum([]byte(""))), nil
}