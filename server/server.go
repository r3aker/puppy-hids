package main

import (
	"github.com/thonsun/puppy-hids/server/action"
	"github.com/thonsun/puppy-hids/server/models"
	"github.com/thonsun/puppy-hids/server/safecheck"
	"github.com/thonsun/puppy-hids/server/utils"
	"context"
	"crypto/tls"
	"errors"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"io/ioutil"
	"time"
)

const authToken string = "67080fc75bb8ee4a168026e5b21bf6fc"

type Watcher int

func (w *Watcher) GetInfo(ctx context.Context,info *action.ComputerInfo,result *action.ClientConfig) error {
	action.ComputerInfoSave(*info)
	config := action.GetAgentConfig(info.IP)
	*result = config
	return nil
}

func (w *Watcher) PutInfo(ctx context.Context,datainfo *action.DataInfo,result *int) error{
	if len(datainfo.Data) == 0{
		return nil
	}
	datainfo.Uptime = time.Now()
	//if datainfo.Type == "connection" {
	//	utils.Info("[+]rpc recieve data:%v",datainfo)
	//}
	utils.Info("[+]rpc recieve data:%v",datainfo)

	err := action.DataInfoSave(*datainfo)// 实时保存在ES(登录login、进程process、文件file，网络连接connection） ｜ 离线保存在mongodb上
	if err != nil{
		utils.Error("Save Data error:%v",err)
	}
	// 初始化 安装HIDS的集群环境监控：观察模式 系统信息收集
	err = action.DataInfoState(*datainfo)
	if err != nil{
		utils.Error("Stat Data error:%v",err)
	}

	//发往告警分析引擎channel
	safecheck.ScanChan <- *datainfo  // 转入分析引擎
	*result = 1
	return nil

}

func auth(ctx context.Context,req *protocol.Message,token string) error  {
	if token == authToken{
		return nil
	}
	return errors.New("invalid token")
}

func init() {
	utils.SetLogLevel(utils.INFO) // 全局开关 是否输出日志信息
	utils.Debug("server config:%#v",models.Config) // models init 从mongodb中获取到 server 配置config,检测规则rules
	ioutil.WriteFile("cert.pem", []byte(models.Config.Cert), 0666)
	ioutil.WriteFile("private.pem", []byte(models.Config.Private), 0666)

	// 启动心跳线程,定时刷新server配置和server检测规则
	go models.Heartbeat()
	// 启动推送任务线程: 指令下发给daemon,socket通讯 上加密保护
	go action.TaskThread()
	// 启动安全检测线程: rule | blacklist | whitelist | audit 告警生成检测引擎
	go safecheck.ScanMonitorThread()
	// 启动客户端健康检测线程
	go safecheck.HealthCheckThread() //agent 离线统计与告警 ｜ 离线原因
	// ES异步写入线程
	go models.InsertThread()
}

func main() {
	//从mongodb 上获取私钥与证书，init 写入文件
	cert, err := tls.LoadX509KeyPair("cert.pem", "private.pem")
	if err != nil {
		utils.Debug("cert error!")
		panic(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	s := server.NewServer(server.WithTLSConfig(config))
	s.AuthFunc = auth
	s.RegisterName("Watcher", new(Watcher), "")
	utils.Debug("RPC Server started")
	err = s.Serve("tcp", ":33433")
	if err != nil {
		utils.Debug("new rpc server error: ",err)
	}
}
