package task

import (
	"fmt"
	"puppy-hids/daemon/common"
	"regexp"
)

// KillProcess 根据进程名结束进程
func KillProcess(processName string) string {
	var data string
	if ok, _ := regexp.MatchString(`^[a-zA-Z0-1\.\-_]+$`, processName); !ok {
		return ""
	}

	data, _ = common.CmdExec(fmt.Sprintf("kill -9 $(pidof %s)", processName))
	return data
}
