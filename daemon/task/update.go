package task

import (
	"github.com/thonsun/puppy-hids/daemon/common"
	"github.com/thonsun/puppy-hids/daemon/install"
	"github.com/thonsun/puppy-hids/daemon/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func agentUpdate(ip string, installPath string, arch string) (bool, error) {
	var err error
	var agentFilePath string
	var md5str string
	agentFilePath = installPath + "agent"
	md5str,err = install.FileMD5String(agentFilePath)
	if err == nil {
		checkURL := fmt.Sprintf("%s://%s/json/check?md5=%s", common.Proto, ip, md5str)
		log.Debug("Start to get url: %s", checkURL)
		res, err := common.HTTPClient.Get(checkURL)
		if err != nil {
			return false, err
		}
		resp, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return false, err
		}
		var result string
		err = json.Unmarshal(resp,&result)
		if err != nil{
			return false,err
		}
		log.Debug("check version result:%v",string(result))
		if string(result) == "1" { // daemon 中agent 版本和 服务器一致，不需要download 进行升级
				return false,nil
		} else {
			common.M.Lock()
			defer common.M.Unlock()
			log.Debug("Updated Agent")
			common.KillAgent()
			time.Sleep(time.Second)
			if err = os.Remove(agentFilePath); err == nil {
				if err = install.DownAgent(ip, agentFilePath, arch); err == nil {
					log.Debug("Download replacement success")
					return true, nil
				}
			return false, err
			}
		}
	}
	return false, err
}
