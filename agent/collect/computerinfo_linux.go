// +build linux

package collect

import (
	"puppy-hids/agent/common"
	"io/ioutil"
	"os"
	"strings"
)

func GetHostInfo() (info common.ComputerInfo) {
	info.IP = common.LocalIP
	info.Hostname, _ = os.Hostname()
	out := common.Cmdexec("uname -r")
	dat, err := ioutil.ReadFile("/etc/redhat-release")
	if err != nil {
		dat, _ = ioutil.ReadFile("/etc/issue")
		issue := strings.SplitN(string(dat), "\n", 2)[0]
		out2 := common.Cmdexec("uname -m")
		info.System = issue + " " + out + out2
	} else {
		info.System = string(dat) + " " + out
	}
	// TODO：web 目录识别，现在由server下发监控文件的变动
	return info
}

func GetAllInfo() AllInfo {
	// 定时收集系统信息：userlist
	// TODO: crontab,listening,loginlog,processlist,service,startup
	allInfo["userlist"] = GetUser()
	return allInfo
}

