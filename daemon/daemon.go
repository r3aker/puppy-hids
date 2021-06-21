package main

//注册daemon进程为系统服务：
//等效于 nohup 的运行，但更为优雅
// daemon -web xxxx:xxx -install

import (
	"flag"
	"os"
	"os/exec"
	"puppy-hids/daemon/task"
	"time"

	"puppy-hids/daemon/common"
	"puppy-hids/daemon/install"
	"puppy-hids/daemon/log"

	"github.com/kardianos/service"
)

var (
	ip            *string
	installBool   bool
	uninstallBool bool
	registerBool  bool
	logfile       *os.File
	err           error
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// 启动tcp server 监听命令连接
	go task.WaitThread() // daemon 真正的程序逻辑
	var agentPath string = common.InstallPath + "agent"
	fileinfo, _ := os.Stat(agentPath)
	if fileinfo.Size() == 0 {
		log.Error("agent error")
		return
	}
	for {
		common.M.Lock() // 锁 控制 后面执行server 指令对agent 启动影响
		// 启动agent
		common.Cmd = exec.Command(agentPath, common.WebIP) // agent 需要管理员对权限 agent
		//err := common.Cmd.Run()
		err := common.Cmd.Start()
		common.M.Unlock()
		if err == nil {
			common.AgentStatus = true
			log.Debug("Start agent successful")
			//阻塞等待agent退出
			err = common.Cmd.Wait()
			if err != nil {
				common.AgentStatus = false
				log.Error("agent exit:%v", err)
			}
		} else {
			log.Debug("Start agent failed:%v", err)
		}
		time.Sleep(time.Second * 30) // 重启agent,sleep减少性能消耗
	}
}

func (p *program) Stop(s service.Service) error {
	common.KillAgent()
	return nil
}

func main() {
	log.SetLogLevel(log.DEBUG)
	err = os.MkdirAll(common.LOG_PATH, os.ModePerm)
	if err != nil {
		log.Error("create log path error:%v", err)
		return
	}
	//TODO:根据日期创建日志文件
	logfile, err = os.OpenFile(common.LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Error("create log file error:%v", err)
		return
	}
	log.SetOutput(logfile)
	flag.StringVar(&common.WebIP, "web", "", "necessary,web ip or domain,eg:xx.xx.xx.xx,security.thonsun.puppyhidsweb.byted.org")
	flag.BoolVar(&installBool, "install", false, "install daemon service")
	flag.BoolVar(&uninstallBool, "uninstall", false, "uninstall daemon service")
	flag.BoolVar(&registerBool, "register", false, "register daemon service")
	flag.Parse()
	if len(os.Args) <= 1 {
		flag.PrintDefaults()
		return
	}
	serviceConfig := &service.Config{
		Name:        "puppy-hids",
		DisplayName: "puppy-hids",
		Description: "实时文件，网络连接，进程监控HIDS",
		Arguments:   []string{"-web", common.WebIP}, // daemon启动绑定web ip ，实现证书，serverlist 下载
	}

	prg := &program{}
	common.Service, err = service.New(prg, serviceConfig)
	if err != nil {
		log.Error("New a service error:%v", err)
		return
	}
	// 卸载服务
	if uninstallBool {
		task.UninstallAll()
		return
	}
	// 安装agent,daemon,并注册daemon 为系统服务
	if installBool {
		if _, err = os.Stat(common.InstallPath); err != nil {
			os.Mkdir(common.InstallPath, 0)
			//检查libpcap是否安装
			err = install.Dependency()
			if err != nil {
				log.Error("puppy-hids dependency error:%v", err)
				return
			}
		}
		if common.WebIP == "" {
			flag.PrintDefaults()
			return
		}

		// 下载agent 并检查文件完整性 ,安装并注册daemon 为系统服务
		err := install.Agent(common.WebIP, common.InstallPath)
		if err != nil {
			log.Error("Install puppy-hids error:%v", err)
		}
		log.Debug("Install puppy-hids service successful")
		return
	}

	// 注册成系统服务
	if registerBool {
		err = common.Service.Install()
		if err != nil {
			log.Error("Install daemon as service error:%v", err)
		} else {
			log.Debug("Install daemon as service success")
			if err = common.Service.Start(); err != nil {
				log.Error("Puppy-hids service start error:%v", err)
			} else {
				log.Debug("Puppy-hids service start success")
			}
		}
		return
	}

	err = common.Service.Run()
	if err != nil {
		// 服务运行的错误
		log.Error("Service run error:%v", err)
	} else {
		log.Debug("Service run successful")
	}
}
