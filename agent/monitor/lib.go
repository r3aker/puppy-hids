package monitor

import (
	"github.com/thonsun/puppy-hids/agent/common"
	"github.com/thonsun/puppy-hids/agent/common/log"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"math"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// 小文件才计算md5
func getFileMD5(path string) (string, error) {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if fileinfo.Size() >= fileSize {
		return "", errors.New("big file")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	md5Ctx := md5.New()
	if _, err = io.Copy(md5Ctx, file); err != nil {
		return "", err
	}
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr), nil
}

func iterationWatcher(monList []string, watcher *fsnotify.Watcher, pathList []string)  {
	for _,p := range monList{
		filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				log.Error("file walk error: %v",err)
				return nil // continue to walk
			}
			if f.IsDir(){
				pathList = append(pathList,path)
				//err = watcher.Add(strings.ToLower(path))
				//log.Debug("add new wather: %v",strings.ToLower(path))
				err = watcher.Add(path)
				log.Debug("add new wather: %v",path)
				if err != nil{
					log.Error("add file watcher error: %v %v",err,path)
				}
			}
			return nil
		})
	}
}

func isFileWhite(resultdata map[string]string) bool {
	for _, v := range common.Config.Filter.File {
		if ok, _ := regexp.MatchString(`^[0-9a-zA-Z]{32}$`, v); ok {// server 端记录hash
			if strings.ToLower(v) == strings.ToLower(resultdata["hash"]) {
				return true
			}
		} else {
			if ok, _ := regexp.MatchString(v, strings.ToLower(resultdata["file"])); ok {
				return true
			} // server端记录 白名单
		}
	}
	return false
}

// 根据远程IP:PORT 获取进程 socketid 宿主主机
func GetSocketID(port int,ip string)(socketID string, err error){
	f, err := ioutil.ReadFile("/proc/net/tcp")
	// 宿主主机读这个文件
	// docker 要遍历/proc/$pid/net/tcp
	// 可以合并为都遍历这个/proc/$pid/net/tcp 类似getpid了
	if err != nil {
		return "", err
	}

	// TODO:socket inode = 0的情况：连续快速 两次curl url，系统没及时更新 local:remte socke_inode的更新
	// 这种用户态能抓的是不全的
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
		fip := IntToIP(uint(ipnum))
		if fip == ip && int(fport)== port{
			socketID = info[9]
			return socketID,nil
		}
	}
	return "",errors.New("cannot find socketid")
}

func GetDockerSocketID(port int, ip string) (socketID string, err error) {
	var pid int
	procDirlist, err := ioutil.ReadDir("/proc")
	if err != nil{
		log.Error("[+]read diraction /proc error:%v",err)
		return "0",errors.New("read diraction /proc error:"+err.Error())
	}
	for _,procDir := range procDirlist {
		if procDir.IsDir() {
			sub := procDir.Name()
			pid, err = strconv.Atoi(sub)
			// 不是数字的将不在遍历范围
			if err != nil {
				log.Debug("not pid:%v", sub)
				continue
			}
			cgroupf := fmt.Sprintf("/proc/%d/cgroup", pid)
			// pid 不存在了进程退出
			cgroups, err:= ioutil.ReadFile(cgroupf)
			if err != nil {
				log.Debug("pid:%v exit",pid)
				continue
			}
			if ok := strings.Contains(string(cgroups[:]),"docker");!ok{
				// 不是docker 的进程
				// TODO:识别docker 进程
				continue
			}
			f, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/net/tcp",pid)) // 短暂进程退出
			if err != nil {
				return "", err
			}

			lines := strings.Split(string(f),"\n")
			for _,line := range lines[1:len(lines)-1]{ // 最后\n
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
				fip := IntToIP(uint(ipnum))
				if fip == ip && int(fport)== port{
					socketID = info[9]
					log.Debug("[+]find docker socket inode:%v",socketID)
					return socketID,nil
				}
			}
		}
	}
	return "",errors.New("cannot find docker socketid")
}
// IntToIP 把数值转为ip字符串
func IntToIP(i uint) string {
	if i > math.MaxUint32 {
		return "beyond the scope of ipv4"
	}

	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)

	return ip.String()
}

// IP string to int
func ipToInt(ip string) int64 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)
	return sum
}

// audit进程创建的信息缓存：map[pid] map[string]string name | path | cmdline
func GeneratePIDInfo(pid int) {
	if _, ok := common.LocalPIDInfo[pid];ok {
		// 已经存在信息不用再次查找
		return
	}
	// 二维map的初始化
	// LRU 在清除pid的时候这个也要清理
	common.LocalPIDInfo[pid] = make(map[string]string)
	common.LocalPIDInfo[pid]["path"] = findPIDPath(strconv.Itoa(pid))
	common.LocalPIDInfo[pid]["name"] = findPIDName(strconv.Itoa(pid))
	common.LocalPIDInfo[pid]["cmdline"] = findPIDCmdline(strconv.Itoa(pid))
	log.Debug("pid %d info:%#v",pid,common.LocalPIDInfo)
}
// audit进程创建时 获取进程打开socket inode 进 socket inode <=> pid 缓存表
// pid 定位所有打开socket inode
// pid ->[socketid...] 这个需要遍历查找到时间，但这个可以控制下面socketid -> pid 表的过期策略删除socket-pid的键值对
// socktid -> pid
func GenerateSocketPIDHash(pid int) error{
	fdDir := fmt.Sprintf("/proc/%d/fd",pid)
	fdSubDirList, err := ioutil.ReadDir(fdDir)
	if err != nil {
		log.Debug("[+]generate hash: read diraction /proc/%d/fd error: %v",pid,err)
		return err
	}
	// 过早读了没有socket的文件，接收到connect 调用开始查询
	var socketInodeArray []string
	for _, socketFile := range fdSubDirList {
		socket := fmt.Sprintf("/proc/%d/fd/%s",pid,socketFile.Name())
		data, err := os.Readlink(socket)
		if err != nil {
			log.Debug("[+]read socket link error:%v",err)
			continue
		}
		log.Debug("[+]link:%v",data)
		if strings.HasPrefix(data,"socket:["){
			socketid := strings.Split(data,"[")[1]
			socketid = socketid[:len(socketid)-1]
			// 保存打开过的socket inode 记录
			// LRU 过期算法实现 控制内存大小
			common.LocalSocketInodePID[socketid] = pid
			//common.LocalPIDSocketInode[pid] = append(common.LocalPIDSocketInode[pid],socketid)
			socketInodeArray = append(socketInodeArray,socketid)
		}
	}// 遍历完成获取PID 进程打开的 socket inode 列表
	//log.Info("ip:%v inode:%v",pid,socketInodeArray)
	// 能够拿到docker 里面的进程 的连接 inode
	// docker 进程 网络连接的识别有别于 宿主主机的
	common.LocalPIDSocketCacheLRU.Set(pid,socketInodeArray)

	//log.Debug("[+]pid:%d socket inode list:%v",pid,common.LocalPIDSocketInode)
	log.Debug("[+]pid %d socket inode:%v",pid,socketInodeArray)
	return nil
}

// 根据pid 识别进程：docker | host
func isDocker(pid string, ppid string) bool {
	cgroupf := fmt.Sprintf("/proc/%s/cgroup", pid)
	// pid 不存在了进程退出
	cgroups, err:= ioutil.ReadFile(cgroupf)
	if err != nil {
		log.Debug("pid:%v exit",pid)
		pcgroupf := fmt.Sprintf("/proc/%s/cgroup", ppid)
		pcgroups, err:= ioutil.ReadFile(pcgroupf)
		if err != nil{
			return false
		}
		if ok := strings.Contains(string(pcgroups[:]),"docker");ok{
			//docker 的进程
			return true
		}
	}
	if ok := strings.Contains(string(cgroups[:]),"docker");ok{
		//docker 的进程
		return true
	}
	return false
}
// 根据socketid 获取进程pid
func GetPID(sockeid string) (pid int,e error) {
	socketInfo := fmt.Sprintf("socket:[%s]",sockeid)
	procDirlist, err := ioutil.ReadDir("/proc")
	if err != nil{
		log.Error("[+]read diraction /proc error:%v",err)
		return -1,errors.New("read diraction /proc error:"+err.Error())
	}
	for _,procDir := range procDirlist{
		if procDir.IsDir(){
			sub := procDir.Name()
			pid, err = strconv.Atoi(sub)
			// 不是数字的将不在遍历范围
			if err != nil {
				log.Debug("not pid:%v",sub)
				continue
			}
			fdDir := fmt.Sprintf("/proc/%d/fd",pid)
			// pid 不存在了进程退出
			fdSubDirList, err := ioutil.ReadDir(fdDir)
			if err != nil {
				log.Error("[+]read diraction /proc/%d/fd error: %v",pid,err)
				continue
			}
			for _, socketFile := range fdSubDirList {
				socket := fmt.Sprintf("/proc/%d/fd/%s",pid,socketFile.Name())
				data, err := os.Readlink(socket)
				if err != nil {
					log.Error("[+]read socket link error:%v",err)
					//continue contine 可能跳过了下一个检查,要继续检查
				}
				if socketInfo == data{
					return pid,nil
				}
			}
		}
	}

	return -1, errors.New("no match socket inodeid")
}

// 根据PID 获取运行路径
func findPIDPath(pid string) (path string){
	var err error
	path, err = filepath.EvalSymlinks(fmt.Sprintf("/proc/%s/cwd",pid))
	if err != nil {
		//TODO:找不到就是空，Error太难看
		log.Debug("[+]open /proc/%v/cwd error:%v",pid,err)
		return ""
	}
	return path
}

// 根据PID 获取进程名
func findPIDName(pid string) (name string)  {
	res, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/comm",pid))
	if err != nil{
		log.Debug("[+]open /proc/%v/comm error:%v",pid,err)
	}
	name = string(res)
	return
}

//根据PID 获取运行 shell 命令具体
func findPIDCmdline(pid string) (cmdline string) {
	res, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/cmdline",pid))
	if err != nil {
		log.Debug("[+]open /proc/%v/cmdline %v",pid,err)
	}
	return string(res)
}

func splitPortService(portService string) (port string) {
	t := strings.Split(portService, "(")
	if len(t) > 0 {
		port = t[0]
	}
	return port
}

func GetFileUID(filename string) (uid uint64) {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		log.Debug("get fileinfo error:%#v",err)
		return
	}
	uid_num := reflect.ValueOf(fileinfo.Sys()).Elem().FieldByName("Uid").Uint()
	return uid_num
}

