package common

import (
	"github.com/thonsun/puppy-hids/daemon/log"
	"crypto/tls"
	"fmt"
	"github.com/kardianos/service"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	// M 安全锁
	M *sync.Mutex
	// Cmd agent进程
	Cmd *exec.Cmd
	// Service daemon服务
	Service service.Service
	// WebIP 服务IP地址
	WebIP string
	// AgentStatus agent状态，是否启动
	AgentStatus bool
	// InstallPath agent安装目录
	InstallPath string
	// Arch 系统位数
	Arch string
	// PublicKey 与Server通讯公钥
	PublicKey string
	// HTTPClient httpclient
	HTTPClient *http.Client
	// Proto 请求协议，测试模式为HTTP
	Proto string
)
//结束agent
func KillAgent() error {
	log.Debug("puppy-hids agent status:%v",AgentStatus)
	if AgentStatus {
		log.Debug("start to kill agent process")
		return Cmd.Process.Kill()
	}
	return nil
}

func init() {
	M = new(sync.Mutex)
	if TESTMODE {
		Proto = "http"
	}else {
		Proto = "https"
	}
	InstallPath = `/usr/puppy-hids/`
	// 系统位数
	if data, _ := CmdExec("getconf LONG_BIT"); InArray([]string{"32", "64"}, data, false) {
		Arch = data
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MaxVersion: 0},
	}
	HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   time.Second * 60,
	}
}


func CmdExec(cmd string) (string, error){
	var c *exec.Cmd
	var data string
	c = exec.Command("/bin/sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return data, err
	}
	data = string(out)

	return data, nil
}

// InArray 判断值是否存在于指定列表中，like为true则为包含判断
func InArray(list []string, value string, like bool) bool {
	for _, v := range list {
		if like {
			if strings.Contains(value, v) {
				return true
			}
		} else {
			if value == v {
				return true
			}
		}
	}
	return false
}

// 获取一个可以绑定的内网IP
func BindAddr() string {
	// 通过连接一个可达的任何一个地址，获取本地的内网的地址,本地调用可能会有多个地址
	log.Debug("req web get local address")
	//conn, err := net.Dial("udp", "10.227.18.247")
	//defer conn.Close()
	//if err != nil {
	//	log.Error("get local ip error:%v",err)
	//}
	//localAddr := conn.LocalAddr().String()
	//log.Debug("[+]Get local address %v",localAddr)
	//idx := strings.LastIndex(localAddr, ":")
	//url := fmt.Sprintf("%s:65512", localAddr[0:idx])
	//log.Debug("bind tcp socket:%v",url)
	var ip string
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/json/getip",WebIP), nil)
	request.Close = true
	if res, err := HTTPClient.Do(request); err == nil {
		defer res.Body.Close()
		result, err := ioutil.ReadAll(res.Body)
		if err != nil{
			return ""
		}
		ip = string(result)
	}
	log.Debug("get local ip:%v",ip)
	return fmt.Sprintf("%s:65512",ip)
}
