// +build linux

package collect

import (
	"github.com/thonsun/puppy-hids/agent/common/log"
	"io/ioutil"
	"strings"
)

// 从/etc/passwd 监控用户变更
func GetUser() (resultData Info) {
	dat, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return
	}
	userList := strings.Split(string(dat), "\n")

	if len(userList) < 2 {
		return
	}
	for _, info := range userList[0 : len(userList)-1] { // 最后一个 \n
		if strings.Contains(info, "/nologin") { // 不允许登陆
			continue
		}
		if strings.Contains(info,"/bin/false") {
			continue
		}
		if strings.Contains(info,"/bin/sync") {
			continue
		}
		s := strings.SplitN(info, ":", 5)// 获取用户uid,gid

		m := map[string]string{"name": s[0],"uid":s[2],"gid":s[3],"description": s[4]}
		log.Debug("get user: %v",m)
		resultData = append(resultData, m)
	}
	log.Debug("user num: %d",len(resultData))
	return
}