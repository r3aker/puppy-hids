// Package collect 获取以下服务器关键信息
// 监听端口，服务列表，用户列表，启动项，计划任务，登录日志
package collect

import (
	"puppy-hids/agent/common"
	"time"
)
type AllInfo map[string][]map[string]string

type Info []map[string]string // PutData中的格式

var allInfo = make(AllInfo)


func init() {
	go func() {
		time.Sleep(time.Second * 3600)
		common.LocalInfo = GetHostInfo()
	}()
}

