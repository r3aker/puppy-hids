package install

import (
	"os"
	"github.com/thonsun/puppy-hids/daemon/common"
	"github.com/thonsun/puppy-hids/daemon/log"
)

func Agent(ip string,installPath string) error {
	log.Debug("Download agent")
	// 下载，安装agent
	err := DownAgent(ip,installPath+"agent",common.Arch)
	if err != nil {
		return err
	}

	// 拷贝自身到安装目录
	log.Debug("Copy the daemon to the installation directory")
	err = copyMe(installPath)
	if err != nil {
		return err
	}
	// 安装daemon为服务,更改daemon 可执行
	os.Chmod(installPath+"daemon", 0750)
	cmd := installPath + "daemon -register -web " + ip
	log.Debug("exec cmd:%s",cmd)
	out, err := common.CmdExec(cmd)
	if err != nil {
		return err
	}
	//启动服务
	log.Debug("start the service...")
	cmd = "systemctl start puppy-hids"// 前面通过第三方包已经把deamon 安装为yulong-hids的服务
	out, err = common.CmdExec(cmd)
	if err == nil && len(out) == 0 {
		log.Debug("Start service successfully")
	}
	return nil
}
