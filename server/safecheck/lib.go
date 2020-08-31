package safecheck

import (
	"github.com/thonsun/puppy-hids/server/utils"
	"regexp"
)

func inArray(list []string, value string, regex bool) bool {
	for _, v := range list {
		if regex {
			if v == "" {
				continue
			}
			if ok, err := regexp.MatchString(v, value); ok {
				return true
			} else if err != nil {
				utils.Debug(err.Error())
			}
		} else {
			if value == v {
				return true
			}
		}
	}
	return false
}

func sendNotice(level int, info string) error {
	utils.Alert("[+]临时告警中心[+] level:%v info:%v",level,info)
	// 告警发送形式
	return nil
	//if models.Config.Notice.Switch {
	//	if models.Config.Notice.OnlyHigh {
	//		if level == 0 {
	//			_, err := http.Get(strings.Replace(models.Config.Notice.API, "{$info}", url.QueryEscape(info), 1))
	//			if err != nil {
	//				return err
	//			}
	//		}
	//		return nil
	//	}
	//	_, err := http.Get(strings.Replace(models.Config.Notice.API, "{$info}", url.QueryEscape(info), 1))
	//	if err != nil {
	//		return err
	//	}
	//}
	//return nil
}