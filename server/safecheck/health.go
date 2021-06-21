package safecheck

import (
	"puppy-hids/server/models"
	"puppy-hids/server/utils"
	"github.com/paulstuart/ping"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"time"
)

// HealthCheckThread 客户端健康检测线程
func HealthCheckThread() {
	utils.Debug("Start Agent Health Check Thread...")
	// TODO:实现对agent 离线原因 查找
	//go offlineCheckThread()
	//go cleanThread()
	//firewallCheckThread()
}

// 离线超过72小时自动清理掉:认为是机器｜容器 已经下线不存在
func cleanThread() {
	client := models.DB.C("client")
	for {
		var offlineIPList []string
		err := client.Find(bson.M{"uptime": bson.M{"$lte": time.Now().Add(time.Hour * time.Duration(-72))}}).Distinct("ip", &offlineIPList)
		if err != nil {
			log.Println("Mongodb query error in cleanThread:", err.Error())
		}
		if len(offlineIPList) >= 100 {
			time.Sleep(time.Second * 60)
			continue
		}
		for _, ip := range offlineIPList {
			err = models.DB.C("client").Remove(bson.M{"ip": ip})
			if err != nil {
				log.Println("Mongodb remove error in cleanThread:", err.Error())
			}
		}

		time.Sleep(time.Second * 60)
	}
}

// offlineCheckThread 离线机器检测 | 超过20台告警
func offlineCheckThread() {
	var oneMinuteAgo time.Time
	var offlineIPList []string
	var msg string
	var cache []string
	client := models.DB.C("client")
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		for _ = range ticker.C {
			cache = []string{} // 一天为一个离线缓存
		}
	}()
	for {
		oneMinuteAgo = time.Now().Add(time.Minute * time.Duration(-5))
		// 通过mongodb找出Agent uptime小于5分钟前的主机到 offlineIPList
		err := client.Find(bson.M{"uptime": bson.M{"$lte": oneMinuteAgo}}).Distinct("ip", &offlineIPList)
		if err != nil {
			log.Println(err.Error())
		}
		// 超过20台掉线直接告警
		if len(offlineIPList) >= 20 {
			err = models.DB.C("notice").Insert(bson.M{"type": "abnormal", "ip": offlineIPList[0], "source": "服务异常", "level": 1,
				"info": offlineIPList[0], "description": "大量主机异常下线，需尽快排查原因。", "status": 0, "time": time.Now()})
			if err == nil {
				msg = fmt.Sprintf("IP:%s,Type:%s,Info:大量主机异常下线，需尽快排查原因。", offlineIPList[0], "abnormal")
				sendNotice(0, msg) // 通知告警
			}
		}
		// 检查离线主机的原因（在小于20台机器）
		for _, ip := range offlineIPList {
			// 健康状态设置为离线 (0健康 1离线 2存在防火墙阻拦)
			client.Update(bson.M{"ip": ip}, bson.M{"$set": bson.M{"health": 1}})

			// 如果开启了离线检测通知才进行ICMP判断并写入警告
			if !models.Config.OfflineCheck || len(offlineIPList) >= 20 {
				continue
			}
			// 机器存活但服务中断5分钟
			if ping.Ping(ip, 3) {
				if inArray(cache, ip, false) {
					continue
				}
				cache = append(cache, ip)
				// 双重告警mongodb插入异常改为sendNotice 通知
				err = models.DB.C("notice").Insert(bson.M{"type": "abnormal", "ip": ip, "source": "服务异常", "level": 1,
					"info": ip, "description": "主机存活但服务未正常工作，可能为被入侵者关闭。", "status": 0, "time": time.Now()})
				if err == nil {
					msg = fmt.Sprintf("IP:%s,Type:%s,Info:主机存活但服务未正常工作，可能为被入侵者关闭。", ip, "abnormal")
					sendNotice(0, msg)
				} else {
					log.Println(err.Error())
				}
			}
		}
		time.Sleep(time.Second * 30)
	}
}

// firewallCheckThread 检测是否通信顺畅 ｜ 在离线的机器集合上 找出哪些是firewall 的问题机器 进行恢复状态（管理人已经处理告警）
func firewallCheckThread() {
	client := models.DB.C("client")
	var onlineIPList []string
	var errIPList []string
	ticker := time.NewTicker(time.Second * 60)
	for _ = range ticker.C {
		client.Find(bson.M{"health": 0}).Distinct("ip", &onlineIPList)
		for _, ip := range onlineIPList {
			conn, err := net.DialTimeout("tcp", ip+":65512", time.Second*3)
			if err != nil {
				client.Update(bson.M{"ip": ip}, bson.M{"$set": bson.M{"health": 2}})
			} else {
				conn.Close()
			}
		}

		// 恢复状态
		client.Find(bson.M{"health": 2}).Distinct("ip", &errIPList)
		for _, ip := range errIPList {
			conn, err := net.DialTimeout("tcp", ip+":65512", time.Second*3)
			if err == nil {
				client.Update(bson.M{"ip": ip}, bson.M{"$set": bson.M{"health": 0}})
				conn.Close()
			}
		}
	}
}