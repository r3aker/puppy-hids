package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	socketid, err := GetSocketID(27017,"192.168.8.114")
	if err != nil {
		panic(err)
	}
	fmt.Println(socketid)
	pid := GetPID(socketid)
	fmt.Println(findPIDPath(pid))
	fmt.Println(findPIDName(pid))
	fmt.Println(findPIDCmdline(pid))
}

func findPIDname(pid string) (name string){
	realPath, err := filepath.EvalSymlinks(fmt.Sprintf("/proc/%s/cwd"))
	if err != nil {
		panic(err)
	}
	fmt.Println("符号链接真实路径:" + realPath)
	return
}

// 根据远程IP:PORT 获取进程 socketid
func GetSocketID(port int,ip string)(socketID string, err error){
	f, err := ioutil.ReadFile("/proc/net/tcp")
	if err != nil {
		return "", err
	}
	//log.Printf("%+v",string(f))
	lines := strings.Split(string(f),"\n")
	for _,line := range lines[1:len(lines)-1]{
		info := strings.Fields(line)
		ipstr := strings.Split(info[2],":")[0]
		portstr := strings.Split(info[2],":")[1]
		fport, err := strconv.ParseUint(portstr,16,32)
		if err != nil{
			panic(err)
		}
		ipnum, err := strconv.ParseUint(ipstr, 16, 32)
		if err != nil{
			panic(err)
		}
		fip, err := Long2IPString(uint(ipnum))
		if err != nil{
			panic(err)
		}
		if fip == ip && int(fport)== port{
			socketID = info[9]
			return socketID,nil
		}
	}
	return "",errors.New("cannot find socketid")
}

// Long2IPString 把数值转为ip字符串
func Long2IPString(i uint) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)

	return ip.String(), nil
}

// 根据socketid 获取进程pid
func GetPID(sockeid string) (pid string) {
	var err error
	socketInfo := fmt.Sprintf("socket:[%s]",sockeid)
	fmt.Println(socketInfo)
	procDirlist, err := ioutil.ReadDir("/proc")
	if err != nil{
		panic(err)
	}
	for _,procDir := range procDirlist{
		if procDir.IsDir(){
			pid = procDir.Name()
			if err != nil {
				continue
			}
			fdDir := fmt.Sprintf("/proc/%s/fd",pid)
			fdSubDirList, err := ioutil.ReadDir(fdDir)
			if err != nil {
				panic(err)
			}
			for _, socketFile := range fdSubDirList {
				socket := fmt.Sprintf("/proc/%s/fd/%s",pid,socketFile.Name())
				data, err := os.Readlink(socket)
				if err != nil {
					continue
				}
				if socketInfo == data{
					return pid
				}
			}
		}
	}
	return ""
}

func findPIDPath(pid string) (path string){
	var err error
	path, err = filepath.EvalSymlinks(fmt.Sprintf("/proc/%s/cwd",pid))
	if err != nil {
		//loginfo("%v",err)
	}
	return
}

func findPIDName(pid string) (name string)  {
	res, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/comm",pid))
	if err != nil{
		//loginfo("%v",err)
	}
	name = string(res)
	return
}

func findPIDCmdline(pid string) (cmdline string) {
	res, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/cmdline",pid))
	if err != nil {
		panic(err)
	}
	return string(res)
}



//func main() {
//	list := []map[string]string{
//		{"name":"thonsun","dsc":"xxxx"},
//		{"name":"test","dsc":"xxxx"},
//		{"name":"test1","dsc":"xxxx"},
//		{"name":"test2","dsc":"xxxx"},
//		{"name":"test3","dsc":"xxxx"},
//		{"name":"test4","dsc":"xxxx"},
//	}
//	fmt.Printf("%+v",list[0:len(list)-2])
//}

//type agent struct {
//	logger *log.Logger
//	debug bool
//}
//func (a *agent)log(format string,info ...interface{})  {
//	if a.debug{
//		a.logger.Printf(format,info...)
//	}
//}
//func main() {
//	client := agent{
//		debug:  true,
//	}
//	client.logger = log.New(os.Stdout,"",log.LstdFlags | log.Lshortfile)
//
//	info := common.ComputerInfo{
//		IP:       "0.0.0.0:193",
//		System:   "ubuntu",
//		Hostname: "hostname",
//		Type:     "linux",
//		Path:     []string{"web path"},
//	}
//
//	client.log("wocao %s,%+v","thonsun",info)
//}

//import (
//	"fmt"
//	"sync"
//	"time"
//)
//
//func main() {
//	wg := sync.WaitGroup{}
//	wg.Add(2)
//	c := make(chan int)
//	go func() {
//		for {
//			time.Sleep(time.Second*4)
//			c <- 1
//		}
//	}()
//
//	go func() {
//		ticker := time.NewTicker(time.Second*1)
//		for {
//			select {
//			case <-ticker.C:
//				fmt.Println("ticker")
//				break
//			case <-c:
//				fmt.Println("another func")
//				break
//			}
//			fmt.Println("for")
//		}
//
//		fmt.Println("func end")
//	}()
//	wg.Wait()
//}

//import (
//	"crypto/tls"
//	"encoding/json"
//	"fmt"
//	"github.com/smallnest/rpcx/client"
//	"io/ioutil"
//	"net/http"
//	"time"
//)
//
//const (
//	AUTH_TOKEN           string          = "67080fc75bb8ee4a168026e5b21bf6fc"
//	CONFIG_REFRESH_INTERVAL int = 60
//	FAIL_MODE client.FailMode = client.Failtry
//	SERVER_API string = "/json/serverlist"
//	TESTMODE bool = true
//	ERROR_LOG_FILENAME = "agent.log"
//)
//
//func main() {
//	servers, err := getServerList("192.168.8.243:8000")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(servers)
//}
//
////从web server 获取当前可用RPC Server
//func getServerList(ip string) ([]string, error) {
//	var server_list []string
//	var url string
//	if TESTMODE{
//		// test_mode 控制web是否使用https
//		//url = "http://" + a.ServerNetLoc + SERVER_API
//		url = "http://" + ip + SERVER_API
//	}else {
//		//url = "https://" + a.ServerNetLoc + SERVER_API
//		url = "https://" + ip + SERVER_API
//	}
//	request, _ := http.NewRequest("GET",url,nil)
//	request.Close = true
//	resp, err := httpClient.Do(request)
//	if err != nil{
//		return nil, err
//	}
//	defer resp.Body.Close()
//	result, err := ioutil.ReadAll(resp.Body)
//	if err != nil{
//		return nil, err
//	}
//	err = json.Unmarshal([]byte(result),&server_list)
//	if err != nil{
//		return nil, err
//	}
//	return server_list,nil
//}
//
//var httpClient = &http.Client{
//	Transport:     &http.Transport{
//		TLSClientConfig:        &tls.Config{
//			InsecureSkipVerify:true, // 使用https，忽略证书验证
//		},
//	},
//	Timeout:       time.Second * 10,
//}

// 什么是BSON
//import (
//"fmt"
//"gopkg.in/mgo.v2/bson"
//)
//
//type Person struct {
//	Name string
//	Phone string ",omitempty"
//}
//
//func main() {
//	data, err := bson.Marshal(&Person{Name:"Bob"})
//	if err != nil {
//		panic(err)
//	}
//	fmt.Printf("%q", data)
//}