// +build linux

package monitor

import (
	"github.com/thonsun/puppy-hids/agent/common"
	"github.com/thonsun/puppy-hids/agent/common/log"
	"github.com/elastic/go-libaudit"
	"github.com/elastic/go-libaudit/auparse"
	"github.com/elastic/go-libaudit/rule"
	"github.com/elastic/go-libaudit/rule/flags"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"
)

func StartProcessMonitor(resultChan chan map[string]string) {
	log.Debug("%s","Start process monitor...")
	common.LocalPIDSocketCacheLRU = common.NewLRUCache(20) // 增加进程LRU 维持cache 大小
	if err := read(resultChan); err != nil {
		log.Error("error: %v", err)
		return
	}
}

func read(resultChan chan map[string]string) error {
	if os.Geteuid() != 0 {
		return errors.New("you must be root to receive audit data")
	}

	var err error
	var client *libaudit.AuditClient

	client, err = libaudit.NewAuditClient(nil)
	if err != nil {
		return errors.Wrap(err, "failed to create audit client")
	}
	defer client.Close()

	// TODO:执行server下发的audit rule
	log.Debug("%s","start to set rules")
	r := "-a always,exit -F arch=b64 -S execve -k proc_create"
	addRule(r,client)
	r = "-a always,exit -F arch=b32 -S execve -k proc_create"
	addRule(r,client)
	// TODO:增加socket 监控 记录进程打开socket 记录
	r = "-a always,exit -F arch=b64 -S connect -k socket_event"
	addRule(r,client)

	status, err := client.GetStatus()
	if err != nil {
		return errors.Wrap(err, "failed to get audit status")
	}
	log.Debug("received audit status= %v", status)

	// 开启Linux内核audit
	if status.Enabled == 0 {
		log.Debug("%s","enabling auditing in the kernel")
		if err = client.SetEnabled(true, libaudit.WaitForReply); err != nil {
			return errors.Wrap(err, "failed to set enabled=true")
		}
	}

	log.Debug("sending message to kernel registering our PID %v the audit daemon", os.Getpid())
	if err = client.SetPID(libaudit.NoWait); err != nil {
		return errors.Wrap(err, "failed to set audit PID")
	}

	return receive(client,resultChan)
}

func receive(r *libaudit.AuditClient,resultChan chan map[string]string) error {
	for {
		rawEvent, err := r.Receive(false)
		if err != nil {
			return errors.Wrap(err, "receive failed")
		}

		// Messages from 1300-2999 are valid audit messages.
		if rawEvent.Type < auparse.AUDIT_USER_AUTH ||
			rawEvent.Type > auparse.AUDIT_LAST_USER_MSG2 {
			continue
		}

		// RawAuditMessage{
		//		Type: auparse.AuditMessageType(msgs[0].Header.Type),
		//		Data: msgs[0].Data,
		//	}
		//fmt.Printf("type=%#v msg=%#v\n", rawEvent.Type, string(rawEvent.Data))

		// 接收kernel audit的消息类型：systemcall
		switch rawEvent.Type {
		case auparse.AUDIT_SYSCALL:
			// logid 是同一时间产生的audit 记录事件
			re,err := regexp.Compile(`audit\((?P<epoch>\d+.\d+):(?P<logid>\d+)\): arch=\w+ syscall=(?P<syscall_id>\d+) success=(?P<success>\w+) exit=?(?P<exit>\-?\d+) .* ppid=(?P<ppid>\d+) pid=(?P<pid>\d+) auid=(?P<auid>\d+) uid=(?P<uid>\d+) gid=(?P<gid>\d+) euid=(?P<euid>\d+) suid=(?P<suid>\d+) .* comm=\"(?P<comm>[-\w_\.]+)\" exe=\"(?P<exe>[-\w\/_\.]+)\" key=\"(?P<key>[-\w\/_\.]+)\".*?`)
			if err != nil {
				log.Debug("%s","error compile regex")
			}

			k1 := re.SubexpNames()
			v := re.FindAllStringSubmatch(string(rawEvent.Data), -1)
			//log.Info("[+]get audit:%v",string(rawEvent.Data))
			if len(v) == 0{
				log.Debug("%s","no match audit")
				continue
			}
			v1 := v[0]

			res := map[string]string{}
			for i := 1;i<len(k1);i++{
				res[k1[i]] = v1[i]
			}
			// 出现socket connect事件:syscall = 42 (一定存在打开的socket inode 文件，此时可以去读取proc)
			// 增加LRU 支持
			if res["syscall_id"] == "42" {
				pid, _ := strconv.Atoi(res["pid"])
				err := GenerateSocketPIDHash(pid)
				if err != nil {
					// 短暂进程找不到
					continue
				}
				GeneratePIDInfo(pid)
				// 这里收集进程的socket 不上报数据
				continue
			}
			//map[string]string{"auid":"4294967295", "comm":"unix_chkpwd", "epoch":"1589164426.375", "euid":"1000", "exe":"/sbin/unix_chkpwd", "exit":"0", "gid":"1000", "pid":"17794", "ppid":"2155", "success":"yes", "suid":"1000", "uid":"1000"}
			res["name"] = res["exe"]
			res["path"] = findPIDPath(res["pid"])// TODO:短暂的进程可能找不到，如ping,容器内进程
			res["cmdline"] = findPIDCmdline(res["pid"]) // TODO:docker 进程识别
			delete(res,"exe")
			res["source"] = "process"
			res["pname"] = findPIDName(res["ppid"])
			res["ppath"] = findPIDPath(res["ppid"])
			res["pcmdline"] = findPIDCmdline(res["ppid"])
			res["process_type"] = "host"
			if isDocker(res["pid"],res["ppid"]) {
				res["process_type"] = "docker"
			}
			resultChan <- res

			// TODO：在进程创建的时候抓取它的socket inode id // 时间消耗 | 过早没有

		case auparse.AUDIT_CONFIG_CHANGE:
			//audit(1589127154.801:4377148): auid=4294967295 ses=4294967295 op=add_rule key="proc_create" list=4 res=1
			re,err := regexp.Compile(`audit\((?P<epoch>\d+.\d+):(?P<logid>\d+)\): auid=(?P<auid>\d+) .+? op=(?P<op>[-\.\w]+) key=\"(?P<key>[-\w\/_\.]+)\".*?`)
			if err != nil {
				log.Debug("%s","error compile regex")
			}

			k1 := re.SubexpNames()
			v := re.FindAllStringSubmatch(string(rawEvent.Data), -1)
			// 注意原来系统的auditd 与 rule 规则要删除，否则匹配不到
			//log.Info("[+]get audit:%v",string(rawEvent.Data))
			if len(v) == 0{
				log.Debug("%s","no match audit")
				continue
			}
			v1 := v[0]
			res := map[string]string{}
			for i := 1;i<len(k1);i++{
				res[k1[i]] = v1[i]
			}
			res["source"] = "audit_config_change"
			resultChan <- res

		//case auparse.AUDIT_SOCKADDR:
			// 一条audit 规则会产生一列 audit 的事件：proctitle,sysaddr,syscall
			// 如何整合 同一事件ID 为一条信息：包含pid ,cmdline, remote addr
			// ausearch -m SYSCALL | -i 对照可以找到audit log的消息解析
			// 0100... local
			// 0200 port ip :remote ip:port
			// 0A00...
			//log.Debug("%s:%v %v","[+]other audit event",rawEvent.Type,string(rawEvent.Data))
		default:
			// TODO:可以定义更多的audit 规则在server audit分析引擎进行告警
			//log.Debug("%s:%v %v","[+]other audit event",rawEvent.Type,string(rawEvent.Data))
		}
	}
}

func addRule(ruleText string,client *libaudit.AuditClient) {
	ruleExec, _ := flags.Parse(ruleText)
	binaryRule,err := rule.Build(ruleExec)
	if err != nil{
		log.Error("failed to build rules:%s %v",ruleText,err)
	}

	if err := client.AddRule(binaryRule); err != nil {
		log.Error("add audit rule err: %v", err)
	}
}
