package action

import (
	"puppy-hids/server/models"
	"puppy-hids/server/utils"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func DataInfoState(datainfo DataInfo) error {
	utils.Debug("agent data log by learn mode statistic")
	var err error
	c := models.DB.C("statistics")
	mainMapping := map[string]string{
		"process":    "name",//data 里面的上报数据,要统计的field值，如进程名，永固名，remote ip名
		"userlist":   "name",
		"listening":  "address",
		"connection": "remote",
		"loginlog":   "remote",
		"startup":    "name",
		"crontab":    "command",
		"service":    "name",
	}
	if _, ok := mainMapping[datainfo.Type]; !ok {
		return nil
	}
	k := mainMapping[datainfo.Type] // infodata.Data 数据包格式中字段的映射 ｜ 找出可以唯一标识的key:ip | user | process
	ip := datainfo.IP
	for _, v := range datainfo.Data {// 找出remote ip地址去除port端口
		if datainfo.Type == "connection" {
			v[k] = strings.Split(v[k], ":")[0]
		}
		// TODO：统计集群实体的维度：进程名，用户名，这个.Count取出会有重复
		count, _ := c.Find(bson.M{"info": v[k], "type": datainfo.Type}).Count()
		if count >= 1 {
			err = c.Update(bson.M{"info": v[k], "type": datainfo.Type}, bson.M{
				"$set":      bson.M{"uptime": datainfo.Uptime},
				"$inc":      bson.M{"count": 1},// coun字段可以作为关闭learn模式后首次出现的规则配置
				"$addToSet": bson.M{"agent_list": ip}})
		} else {
			serverList := []string{ip}
			err = c.Insert(bson.M{"type": datainfo.Type, "info": v[k], "count": 1,// rule 规则里面可以找出这些 =1 的新出现的进程名，用户名
				"agent_list": serverList, "uptime": datainfo.Uptime})
		}
	}
	return err
}