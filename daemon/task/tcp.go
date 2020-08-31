package task

import (
	"bufio"
	"github.com/thonsun/puppy-hids/daemon/common"
	"github.com/thonsun/puppy-hids/daemon/log"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"strings"
)

type taskServer struct {
	TCPListener net.Listener
	ServerIP    string
	ServerList  []string
}

func (t *taskServer) listen() (err error) {
	// 采用http get 获取server 的公钥 进行双方命令，安全在于 http -> 客户端认证http的证书
	log.Debug("daemon bind socket:%v",common.BindAddr())
	t.TCPListener, err = net.Listen("tcp", common.BindAddr())
	return err
}

func (t *taskServer) run() {
	err := t.listen()
	if err != nil {
		return
	}
	for {
		tcpConn, err := t.TCPListener.Accept()
		if err != nil {
			log.Error("Accept new TCP listener error:", err.Error())
			continue
		}
		t.ServerIP = strings.SplitN(tcpConn.RemoteAddr().String(), ":", 2)[0]
		// 只接受serverlist 列表中的tcp socket 连接
		if t.isServer() {
			t.tcpPipe(tcpConn) // 公钥解密 tcp 数据, 处理流程
		} else {
			tcpConn.Close()
		}
	}
}
func (t *taskServer) isServer() bool {
	t.setServerList()
	for _, ip := range t.ServerList {
		if t.ServerIP == strings.SplitN(ip, ":", 2)[0] {
			return true
		}
	}
	return false
}
func (t *taskServer) setServerList() error {
	resp, err := common.HTTPClient.Get(common.Proto + "://" + common.WebIP + common.SERVER_LIST_API)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	json.Unmarshal([]byte(result), &t.ServerList)
	return nil
}

func (t *taskServer) tcpPipe(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	message, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}
	decodeBytes, _ := base64.RawStdEncoding.DecodeString(string(message))// 防止 加密 byte 不能正确传输
	decryptdata, err := rsaDecrypt(decodeBytes)
	if err != nil {
		log.Debug("Decrypt rsa text in tcpPipe error:", err.Error())
		return
	}
	var taskData map[string]string
	err = json.Unmarshal(decryptdata, &taskData)
	if err != nil {
		log.Debug("Unmarshal json text in tcpPipe error", err.Error())
		return
	}
	var taskType string
	var data string
	if _, ok := taskData["type"]; ok {
		taskType = taskData["type"]
	}
	if _, ok := taskData["command"]; ok {
		data = taskData["command"]
	}
	result := map[string]string{"status": "false", "data": ""}
	T := Task{taskType, data, result}
	log.Debug("[+]recieve task:%v %v",T.Type,T.Command)
	if sendResult := T.Run(); len(sendResult) != 0 {
		conn.Write(sendResult)
	}
}

// WaitThread 接收任务线程
func WaitThread() {
	setPublicKey()
	var t taskServer
	t.run()
}
