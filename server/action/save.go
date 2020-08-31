package action

import (
	"github.com/thonsun/puppy-hids/server/models"
	"github.com/thonsun/puppy-hids/server/utils"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"time"
)


// Agent 系统信息
type ComputerInfo struct {
	IP string
	System string
	Hostname string
	Type string
	Path []string

	Uptime time.Time
}

// agent 上报监控数据
type DataInfo struct {
	IP     string `bson:"ip"`
	Type   string `bson:"type"`
	System string `bson:"system"`
	Data   []map[string]string `bson:"data"`
	Uptime time.Time `bson:"uptime"`
}

//save agent info to mongodb agent存活统计
func ComputerInfoSave(info ComputerInfo) error {
	utils.Debug("Agent Info save: %v",info)

	// 保存HIDS Agent IP存活时间
	c := models.DB.C("client")
	info.Uptime = time.Now()
	// where ip=info.ip
	c.Upsert(bson.M{"ip": info.IP}, bson.M{"$set": &info})
	// where ip=info.ip and (health=1 | health=nil)
	c.Update(bson.M{"ip": info.IP, "$or": []bson.M{bson.M{"health": 1}, bson.M{"health": nil}}}, bson.M{"$set": bson.M{"health": 0}})
	return nil
}

func DataInfoSave(datainfo DataInfo) error {
	var err error
	// 登录日志、网络连接、进程创建、文件操作 存放在es，其余保存在mongodb
	if datainfo.Type == "loginlog" || datainfo.Type == "connection" || datainfo.Type == "process" || datainfo.Type == "file" {
		utils.Debug("agent data log to es")
		if datainfo.Type == "loginlog" {
			for _, logininfo := range datainfo.Data {
				time, _ := time.Parse("2006-01-02T15:04:05Z07:00", logininfo["time"])
				delete(logininfo, "time")
				esdata := models.ESSave{
					IP:   datainfo.IP,
					Data: logininfo,
					Time: time,
					DataType:datainfo.Type,
				}
				models.InsertEs(esdata)
			}
		} else {
			dataTimeInt, err := strconv.Atoi(datainfo.Data[0]["time"])
			if err != nil {
				return err
			}
			delete(datainfo.Data[0], "time")
			esdata := models.ESSave{
				IP:   datainfo.IP,
				Data: datainfo.Data[0],
				Time: time.Unix(int64(dataTimeInt), 0),
				DataType:datainfo.Type,
			}
			models.InsertEs(esdata)
		}
	} else {
		utils.Debug("agent data log to mongodb:")
		c := models.DB.C("info")
		count, _ := c.Find(bson.M{"ip": datainfo.IP, "type": datainfo.Type}).Count()//TODO： 这个日志替换不太好，应该全部保存
		if count >= 1 {
			err = c.Update(bson.M{"ip": datainfo.IP, "type": datainfo.Type},
				bson.M{"$set": bson.M{"data": datainfo.Data, "uptime": datainfo.Uptime}})
		} else {
			err = c.Insert(&datainfo)
		}
		return err
	}
	return nil
}

