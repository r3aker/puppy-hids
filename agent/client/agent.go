package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"
	"io/ioutil"
	"net"
	"net/http"
	"puppy-hids/agent/collect"
	"puppy-hids/agent/common"
	"puppy-hids/agent/common/log"
	"puppy-hids/agent/config"
	"puppy-hids/agent/monitor"
	"runtime"
	"strings"
	"sync"
	"time"
)
var err error

var httpClient = &http.Client{
	Transport:     &http.Transport{
		TLSClientConfig:        &tls.Config{
			InsecureSkipVerify:true, // 使用https，忽略证书验证
		},
	},
	Timeout:       time.Second * 10,
}

// Agent 上报数据类型
type dataInfo struct {
	IP string
	Type string //数据类型
	System string //os类型
	Data []map[string]string
}

type Agent struct {
	ServerNetLoc string // web:rest api web server，暂无前端展示
	ServerList   []string // rpc server 可用列表 ip:port
	Client       client.XClient //RPC 客户端
	PutData      dataInfo //RPC 要上报的数据
	Reply        int // RPC server response
	Mutex        *sync.Mutex // 保证在更新 agent 配置安全
	ctx          context.Context
}


//初始化 agent rpc client | rpc call to get agent config
func (a *Agent) init() {
	a.ServerList, err = a.getServerList()
	if err != nil{
		log.Error("get available server list failed: %v",err)
		panic(1)
	}
	log.Debug("Available Server list: %v",a.ServerList)
	a.ctx = context.WithValue(context.Background(),share.ReqMetaDataKey,make(map[string]string))
	if len(a.ServerList) == 0{
		log.Error("%s","No server node available")
		time.Sleep(time.Second * 30) // 减少agent 占用CPU
		panic(1)
	}

	// 初始化RPC client:从可用server 节点选择｜tls
	a.newClient()
	if common.LocalIP == ""{
		log.Error("%s","Cannot get local address")
		panic(1)
	}

	a.Mutex = new(sync.Mutex) // 保证在更新 agent 配置安全
	// 获取agent 配置：监控目录，filter白名单
	err := a.Client.Call(a.ctx,"GetInfo",&common.LocalInfo,&common.Config)
	if err != nil {
		log.Error("RPC Client call errors: %v",err)
		panic(1)
	}
	log.Debug("Common client config: %#v",common.Config)
}

//初始化hids-agent 连接的server的rpc client
func (a *Agent) newClient() {
	var servers []*client.KVPair
	for _, server := range a.ServerList{
		common.ServerIPList = append(common.ServerIPList,strings.Split(server,":")[0])
		s := client.KVPair{Key:server}
		servers = append(servers,&s)// 指针传递，修改对象
		if common.LocalIP == "" {
			a.setLocalIP(server)
			common.LocalInfo = collect.GetHostInfo()
			log.Debug("Host Information: %#v",common.LocalInfo)
		}
	}
	// 忽略server 认证
	conf := &tls.Config{
		InsecureSkipVerify:true,
	}
	option := client.DefaultOption
	option.TLSConfig = conf
	serverd := client.NewMultipleServersDiscovery(servers)
	a.Client = client.NewXClient("Watcher",config.FAIL_MODE,client.RandomSelect,serverd,option)
	a.Client.Auth(config.AUTH_TOKEN)
}

func (a *Agent) configRefresh() {
	ticker := time.NewTicker(time.Second * time.Duration(config.CONFIG_REFRESH_INTERVAL))
	go func() {
		for _ = range ticker.C{
			ch := make(chan bool)
			//定时刷新监控配置：RPC call to get new monitor config,如监控目录变更
			go func() {
				err = a.Client.Call(a.ctx,"GetInfo",&common.LocalInfo,&common.Config)
				if err != nil {
					log.Error("tick to get client config err: %v",err)
					return
				}
				log.Debug("tick to get client config: %#v",common.Config)
				ch <- true // ch 同步控制
			}()

			//server 集群更新
			select {
			case <-ch:
				serverList, err := a.getServerList()
				if err != nil {
					log.Error("get available server list failed: %v", err)
					break
				}
				if len(serverList) == 0 {
					log.Error("%s","no server node available")
					break
				}
				if len(serverList) == len(a.ServerList) {
					for i, server := range serverList {
						if server != a.ServerList[i] {
							log.Info("[+]change server list:%v",serverList)
							a.ServerList = serverList
							// 防止正在传输重置client导致数据丢失
							a.Mutex.Lock()
							a.Client.Close()
							a.newClient()
							a.Mutex.Unlock()
							break
						}
					}
				} else {
					log.Debug("server nodes from old to new:", a.ServerList, "->", serverList)
					a.ServerList = serverList
					a.Mutex.Lock()
					a.Client.Close()
					a.newClient()
					a.Mutex.Unlock()
				}
			case <-time.NewTicker(time.Second * 3).C:
				break
			}
		}
	}()
}

func (a *Agent) monitor() {
	resultChan := make(chan map[string]string,16)
	go monitor.StartNetSniff(resultChan) // google gopacket 获取网卡handler
	go monitor.StartProcessMonitor(resultChan) // go-libaudit 接收内核通知
	go monitor.StartFileMonitor(resultChan) // 上面的配置更新了没有及时的通知到 这些monitor 需要重启agent

	go func(result chan map[string]string) {
		var resultdata []map[string]string
		var data map[string]string
		for {
			data = <- result
			data["time"] = fmt.Sprintf("%d",time.Now().Unix())
			log.Debug("monitor data: %#v",data)
			//小功能测试信息输出
			//log.Info("Monitor data: %#v",data)
			source := data["source"]// 复用channel source 为哪个monitor发来的数据：file | process | connection
			delete(data,"source")
			a.Mutex.Lock()
			a.PutData = dataInfo{
				IP:     common.LocalIP,
				Type:   source,
				System: runtime.GOOS,
				Data:   append(resultdata,data),// 与下面的系统信息传输一致 PutData 格式统一是Info []map[string]string 格式
			}
			a.put()
			a.Mutex.Unlock()
		}
	}(resultChan)

}

func (a *Agent) getInfo() {
	historyCache := make(map[string][]map[string]string)
	for {
		if len(common.Config.MonitorPath) == 0{
			log.Debug("%s","Failed to get the configuration information")
			time.Sleep(time.Second * 5) // 减少Agent 占用cpu资源
			continue
		}
		allData := collect.GetAllInfo() //TODO: map[string] []map[string]string userlist | crontab | service | startup ...
		for k, v := range allData{
			if len(v) == 0 || a.mapComparision(v,historyCache[k]){
				log.Debug("getinfo data: %v %s",k," no change")
				continue
			}else {
				a.Mutex.Lock() // 发送消息要同步控制
				a.PutData = dataInfo{
					IP:     common.LocalIP,
					Type:   k,
					System: runtime.GOOS,
					Data:   v,
				}
				a.put()
				a.Mutex.Unlock()
				log.Debug("Data details: %#v %#v",k, a.PutData)
				historyCache[k] = v
			}
		}
		if common.Config.Cycle == 0 {
			common.Config.Cycle = 1
		}
		time.Sleep(time.Second * time.Duration(common.Config.Cycle) * 60)
	}
}

func (a *Agent) Run() {
	// 请求Web API，获取Server地址，初始化RPC客户端，RPC call 获取初始监控配置
	a.init()

	// time.Tick 定时更新监控配置配置
	a.configRefresh()

	// 启动监控流程：文件监控,网络连接，进程创建
	a.monitor()

	// time.tick 定时获取系统信息
	a.getInfo() // agent.Run 的保证main 线程一直运行 waitgroup使用也可以
}

//从web server 获取当前可用RPC Server
func (a Agent) getServerList() ([]string, error) {
	var server_list []string
	var url string
	if config.TESTMODE{
		// test_mode 控制web是否使用https
		url = "http://" + a.ServerNetLoc + config.SERVER_API
	}else {
		url = "https://" + a.ServerNetLoc + config.SERVER_API
	}
	log.Debug("Get Web RESTAPI Available server: %v",url)
	request, _ := http.NewRequest("GET",url,nil)
	request.Close = true
	resp, err := httpClient.Do(request)
	if err != nil{
		return nil, err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}
	err = json.Unmarshal([]byte(result),&server_list)
	if err != nil{
		return nil, err
	}
	return server_list,nil
}

// tcp连接设置agent IP
func (a Agent) setLocalIP(ip string) {
	conn, err := net.Dial("tcp",ip)
	if err != nil{
		log.Debug("Net.Dial:%v",ip)
		log.Error("error:%v ",err)
		panic(1)
	}
	defer conn.Close()
	common.LocalIP = strings.Split(conn.LocalAddr().String(),":")[0]
}

func (a Agent) mapComparision(new []map[string]string, old []map[string]string) bool {
	if len(new) == len(old){// 定义不同：相同长度再细致比较
		for i, v := range new {
			for k,value := range v{
				if value != old[i][k]{// array i的顺序内容不变
					return false
				}
			}
		}
		return true
	}
	return false
}

func (a Agent) put() {
	// 异步RPC 上传监控信息
	//log.Info("RPC put data:%#v",a.PutData)
	if a.PutData.Type == "connection" {
		log.Info("[+]agent data:%v",a.PutData)
	}
	_, err := a.Client.Go(a.ctx,"PutInfo",&a.PutData,&a.Reply,nil)
	if err != nil {
		//log.Error("Agent PutInfo error: %v",err)
	}
}

