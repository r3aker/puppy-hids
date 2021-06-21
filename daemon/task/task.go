package task

import (
	"encoding/json"
	"puppy-hids/daemon/common"
	"puppy-hids/daemon/log"
	"syscall"
)

// task 接收server 指令
type Task struct {
	Type    string
	Command string
	Result  map[string]string
}

func (t *Task) Run() []byte {
	switch t.Type {
	case "uninstall": // 卸载puppy-hids
		t.uninstall()
	case "update": // 更新agent 版本
		t.update()
	case "reload": // 重新运行puppy-hids
		t.reload()
	case "stop": // 暂停 puppy-hids agent
		t.stop()
	case "continue":
		t.continve()
	case "kill": // 以command 为进程名，kill
		t.kill()
	case "exec": // shell 命令执行 command 要执行的命令
		t.exec()
	}

	var sendResult []byte
	if b, err := json.Marshal(t.Result); err == nil {
		msg := string(b) + "\n"
		sendResult = []byte(msg)
	}
	return sendResult
}

func (t *Task) reload() {
	t.Result["status"] = "true"
	t.Result["data"] = "reload puppy-hids agent successful"
	log.Debug("task reload puppy-hids agent")
	// 服务的for 循环会重启agent
	// TODO:重启有问题
	if err := common.KillAgent(); err != nil {
		log.Debug("reload error:%v", err)
		t.Result["status"] = "false"
		t.Result["data"] = err.Error()
	}
}
func (t *Task) stop() {
	//这个退出应该是退去agent 监控，实施agent 性能降级
	t.Result["status"] = "true"
	t.Result["data"] = "stop agent successful"
	log.Debug("task stop:agent status %v", common.AgentStatus)
	if common.AgentStatus {
		common.Cmd.Process.Signal(syscall.SIGSTOP)
	}
	log.Debug("stop puppy-hids agent by signal")
}

func (t *Task) continve() {
	t.Result["status"] = "true"
	t.Result["data"] = "continue agent successful"
	log.Debug("task continue:agent status %v", common.AgentStatus)
	if common.AgentStatus {
		common.Cmd.Process.Signal(syscall.SIGCONT)
	}
	log.Debug("continue puppy-hids agent by signal")
}
func (t *Task) kill() {
	// 可以杀死主机任意进程名
	// 已经成功kill 没有返回数据
	log.Debug("task kill process:%v", t.Command)
	KillProcess(t.Command)
	t.Result["status"] = "true"
	t.Result["data"] = "ok"

}
func (t *Task) uninstall() {
	log.Debug("task uninstall puppy-hids service")
	t.Result["status"] = "true"
	t.Result["data"] = "uninstall puppy-hids service successful"
	// 停止agent 进程
	UninstallAll()
}
func (t *Task) update() {
	log.Debug("task update agent")
	if ok, err := agentUpdate(common.WebIP, common.InstallPath, common.Arch); err == nil {
		// ok 是需不需要尽心下载更新
		if ok {
			t.Result["status"] = "true"
			t.Result["data"] = "更新完毕"
		} else {
			t.Result["status"] = "true"
			t.Result["data"] = "已经是最新版本" // 依据md5进行判断
		}
	} else {
		t.Result["data"] = err.Error()
	}
}

func (t *Task) exec() {
	log.Debug("task exec:%v", t.Command)
	if dat, err := common.CmdExec(t.Command); err == nil {
		t.Result["status"] = "true"
		t.Result["data"] = dat
	} else {
		t.Result["data"] = err.Error()
	}
}
