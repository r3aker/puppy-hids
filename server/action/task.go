package action

import (
	"bufio"
	"puppy-hids/server/models"
	"puppy-hids/server/utils"
	"encoding/base64"
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net"
	"time"
)

// 保存到mongodb 中的数据格式
// daemon 进程要执行的指令
type queue struct {
	ID bson.ObjectId `bson:"_id"`
	IP string `bson:"ip"`
	Type string `bson:"type"`
	Command string `bson:"command"`
	Time time.Time `bson:"time"`
}
// daemon 指令执行的结果socket 返回
type taskResult struct {
	IP string `bson:"ip"`
	Status string `bson:"status" json:"status"`
	Data string `bson:"data" json:"data"`
	Time time.Time `bson:"time"`
}

var threadpool chan bool

func TaskThread() {
	utils.Debug("Start Server2Agent Task Push Thread...")
	threadpool = make(chan bool,100) // 100线程限制,减少server daemon tcp连接
	for {
		res := queue{}
		change := mgo.Change{
			Remove:    true,// 查询取出并删除
		}
		// 找到自己的命令 ip 标识
		models.DB.C("queue").Find(bson.M{}).Limit(1).Apply(change,&res)
		if res.IP == "" {
			time.Sleep(time.Second * 10) //暂无任务
			continue
		}
		threadpool <- true
		go sendTask(res,threadpool)
	}
}

func sendTask(task queue,threadpool chan bool) {
	defer func() {
		<- threadpool
	}()
	// daemon接收的格式是json marshal结果
	sendData := map[string]string{"type":task.Type,"command":task.Command}
	if data, err := json.Marshal(sendData);err == nil{
		conn, err := net.DialTimeout("tcp",task.IP+":65512",time.Second*3)
		utils.Info("[+]sendtask: %s %s %s",task.IP,task.Command,task.Type)
		if err != nil {
			saveError(task,err.Error())// 保存指令下发的错误
			return
		}
		defer conn.Close()
		encryptData, err := rsaEncrypt(data)
		if err != nil {
			saveError(task,err.Error())
			return
		}
		conn.Write([]byte(base64.RawStdEncoding.EncodeToString(encryptData) + "\n"))
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')
		if err != nil || len(msg) == 0 {
			saveError(task,err.Error())
			return
		}

		utils.Info("[+]recieve msg: %v %v",conn.RemoteAddr().String(),msg)
		res := taskResult{}
		err = json.Unmarshal([]byte(msg),&res)
		if err != nil {
			saveError(task,err.Error())
			return
		}
		res.Time = time.Now()
		res.IP = task.IP
		c := models.DB.C("task_result")
		err = c.Insert(&res)
		if err != nil {
			saveError(task,err.Error())
			return
		}
	}
}

func saveError(task queue,errMsg string){
	utils.Debug("recieve error:%v %#v",errMsg,task)
	res := taskResult{
		IP:     task.IP,
		Status: "false",
		Data:   errMsg,
		Time:   time.Now(),
	}
	c := models.DB.C("task_result")
	err := c.Insert(&res)
	if err != nil {
		utils.Error("insert to mongo error:%v",err)
	}
}
