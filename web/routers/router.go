package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"puppy-hids/web/common"
	"puppy-hids/web/common/log"
	"puppy-hids/web/setting"
	"time"
)

const (
	HOME string = "192.168.8.243:33433"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	gin.SetMode(setting.RunMode)

	// 返回server 公钥
	r.GET("/json/publickey", func(context *gin.Context) {
		var res common.Config
		err := common.DB.C("config").Find(bson.M{"type":"server"}).One(&res)
		if err != nil{
			log.Error("find public key error:%v",err)
		}
		dic := res.Dic.(bson.M)

		context.JSON(200,map[string]string{"public":dic["publickey"].(string)})
	})

	// agent md5校验
	r.GET("json/check", func(context *gin.Context) {
		reqstr := context.Query("md5")
		log.Debug("recieve md5 string:%v",reqstr)
		file := "upload/agent"
		md5str,err := common.FileMD5String(file)
		log.Debug("agent md5 string:%v",md5str)
		if err != nil {
			log.Error("get file md5 error:%v",err)
		}
		var res string
		if reqstr == md5str {
			res = "1"
		}else {
			res = "not match"
		}
		context.JSON(200,res)
	})

	// agent 二进制文件下载
	r.GET("/json/download", func(context *gin.Context) {
		content,err := ioutil.ReadFile("upload/agent")
		if err != nil {
			log.Debug("read file error:%v",err)
		}
		context.Writer.WriteHeader(http.StatusOK)
		context.Header("Content-Disposition", "attachment; filename=agent")
		context.Header("Content-Type", "application/text/plain")
		context.Header("Content-Length", fmt.Sprintf("%d", len(content)))
		context.Writer.Write(content)
	})

	// server 存活列表
	r.GET("/json/serverlist", func(context *gin.Context) {
		// TODO:从mongodb 获取注册的IP 信息组成数组返回
		servers := []common.Server{}
		available_server := []string{}
		err := common.DB.C("server").Find(nil).All(&servers)
		if err != nil {
			log.Error("no available server node")
		}
		for _, server := range servers{
			//log.Printf("%s time:%v\n",server.Netloc,server.Uptime.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
			//log.Printf("time.Now().Sub(server.Uptime):%v",time.Now().Sub(server.Uptime).Minutes())
			if (time.Now().Sub(server.Uptime).Minutes())/10 <  1{ // 10min 之内判断server 是存活的
				available_server = append(available_server,server.Netloc)
			}
		}
		//log.Printf("%v\n",available_server)

		context.JSON(200,available_server)
	})

	// 上传 agent
	r.POST("/json/upload", func(context *gin.Context) {
		// FormFile方法会读取参数“upload”后面的文件名，返回值是一个File指针，和一个FileHeader指针，和一个err错误。
		file, header, err := context.Request.FormFile("upload")
		if err != nil {
			context.String(http.StatusBadRequest, "Bad request")
			return
		}
		// header调用Filename方法，就可以得到文件名
		filename := header.Filename
		fmt.Println(file, err, filename)

		// 创建一个文件，文件名为filename，这里的返回值out也是一个File指针
		err = os.MkdirAll("upload",os.ModePerm)
		if err != nil{
			log.Error("upload create path error:%v",err)
		}
		//out, err := os.OpenFile("upload/"+filename,os.O_TRUNC | os.O_CREATE |os.O_RDWR,0666)
		//if err != nil {
		//	log.Error("upload create *File error:%v",err)
		//}
		out, err := os.Create("upload/"+filename)
		if err != nil {
			log.Error("upload create *File error:%v",err)
		}
		defer out.Close()

		// 将file的内容拷贝到out
		_, err = io.Copy(out, file)
		if err != nil {
			log.Error("upload write file error:%v",err)
		}

		context.String(http.StatusCreated, "upload successful \n")
	})

	// 告警列表获取
	r.GET("json/alert", func(context *gin.Context) {
		var notices []common.Notice
		err := common.DB.C("notice").Find(nil).Sort("-time").Limit(5).All(&notices)
		if err != nil {
			log.Error("find mongodb notice error:%v",err)
		}
		for _,notice := range notices {
			cst, _ := time.LoadLocation("Asia/Shanghai")
			notice.Time = notice.Time.In(cst)
		}
		context.JSON(http.StatusOK,notices)
	})
	// 返回客户端局域网IP
	r.GET("json/getip", func(context *gin.Context) {
		ip := context.ClientIP()
		context.String(http.StatusOK,ip)
	})

	// 获取当前在线agent
	r.GET("json/client", func(context *gin.Context) {
		var clients []common.Client
		var available_clients []common.Client
		err := common.DB.C("client").Find(nil).All(&clients)
		if err != nil{
			log.Error("no alive agent")
		}
		// 离线判断
		for _, client := range clients{
			if time.Now().Sub(client.Uptime).Minutes()/30 < 1{
				cst, _ := time.LoadLocation("Asia/Shanghai")
				client.Uptime = client.Uptime.In(cst)
				available_clients = append(available_clients,client)
			}
		}
		context.JSON(http.StatusOK,available_clients)
	})

	return r
}
