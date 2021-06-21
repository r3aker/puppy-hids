package task

import (
	"os"
	"puppy-hids/daemon/common"
	"puppy-hids/daemon/log"
	"strings"
)

func UninstallAll() {
	common.KillAgent()
	// 注销deamon 服务:如何正确注销服务
	data, err := common.CmdExec("cat /etc/os-release")
	if err != nil {
		log.Error("exec cmd error:%v", err)
	}
	//log.Debug("cmd result:%v",data)
	log.Debug("%v", strings.Contains(data, "debian"))
	if strings.Contains(data, "debian") {
		data, err := common.CmdExec("sudo service puppy-hids stop")
		if err != nil {
			log.Debug("exec error sudo service puppy-hids stop")
		}
		log.Debug("exec data:%v", data)
		data, err = common.CmdExec("sudo update-rc.d puppy-hids remove")
		if err != nil {
			log.Debug("exec error sudo update-rc.d puppy-hids remove")
		}
		log.Debug("exec data:%v", data)
		data, err = common.CmdExec("sudo rm /etc/systemd/system/puppy-hids.service")
		if err != nil {
			log.Debug("exec error sudo rm /etc/systemd/system/puppy-hids.service")
		}
		log.Debug("exec data:%v", data)
		data, err = common.CmdExec("sudo systemctl daemon-reload")
		if err != nil {
			log.Debug("exec error sudo systemctl daemon-reload")
		}
		log.Debug("exec data:%v", data)
	}
	os.Exit(1)
	//if err := common.Service.Uninstall();err != nil {
	//	log.Error("uninstall puppy-hids service error:%v",err)
	//}
}
