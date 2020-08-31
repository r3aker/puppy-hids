package common

import (
	"log"
	"os/exec"
	"regexp"
	"strings"
)

// ClientConfig agent配置，定时从 server 更新client 配置（IP）
type ClientConfig struct {
	Cycle  int    // 信息传输频率，单位：分钟
	UDP    bool   // 是否记录UDP请求
	LAN    bool   // 是否本地网络请求
	Type   string // 主机的类型 web | db | only-linux-server
	Filter struct {
		File    []string // 文件hash、文件名
		IP      []string // IP地址
		Process []string // 进程名
		Port 	[]int // 端口
	} // 直接过滤不回传的信息 ｜ 正则匹配（如文件后缀名）
	MonitorPath []string // 监控目录列表
	Lasttime    string   // 最后一条登录日志时间
}

// ComputerInfo 计算机信息结构
type ComputerInfo struct {
	IP       string   // IP地址
	System   string   // 操作系统
	Hostname string   // 计算机名
	Type     string   // 服务器类型
	Path     []string //TODO:agent自动识别web目录，现在本版本为web配置
}

var (
	//TODO:全局共享变量的并发安全

	// Config 配置信息
	Config ClientConfig
	// LocalIP 本机活跃IP
	LocalIP string
	// AgentInfo 主机相关信息
	LocalInfo ComputerInfo
	// ServerIPList 服务端列表
	ServerIPList []string

	// 记录本地 PID -> [socketid...] 映射：
	// 对于瞬时结束的进程在查询时候决定是否删除
	// LRU 支持：维持一定数量的PID 映射，当队列满的时候，删除最少使用的进程PID过程：
	// 1、从struct 中获取到pid,[socketid...]
	// 2、遍历socketid 删除socketid -> pid map
	// 3、删除pidinfo struct: cmdline,comm,path
	// 添加一个新的进程进入的时候也是如此：
	// 1、LRU Cache 建立 pid->[socketid ...]
	// 2、socketid -> pid
	// 3、pid info struct建立
	LocalPIDSocketCacheLRU *LRUCache
	//LocalPIDSocketInode = make(map[int][]string)

	// 记录socket -> pid 的映射
	LocalSocketInodePID = make(map[string]int)

	// 记录本地 PID name|path|cmdline 的映射
	LocalPIDInfo = make(map[int]map[string]string)
)

// Cmdexec 执行系统命令
func Cmdexec(cmd string) string {
	var c *exec.Cmd
	var data string
	argArray := strings.Split(cmd, " ")
	c = exec.Command(argArray[0], argArray[1:]...)
	out, _ := c.CombinedOutput()
	data = string(out)
	return data
}

// InArray 判断是否存在列表中，如果regex为true，则进行正则匹配
func InArray(list []string, value string, regex bool) bool {
	for _, v := range list {
		if regex {
			if ok, err := regexp.Match(v, []byte(value)); ok {
				return true
			} else if err != nil {
				log.Println(err.Error())
			}
		} else {
			if value == v {
				return true
			}
		}
	}
	return false
}
