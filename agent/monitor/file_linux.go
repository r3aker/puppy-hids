// +build linux

package monitor

import (
	"github.com/thonsun/puppy-hids/agent/common"
	"github.com/thonsun/puppy-hids/agent/common/log"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Watcher *fsnotify.Watcher
	err      error
)

// 根据uid 获取user
func getFileUser(path string) (string, error) {
	uidStr := fmt.Sprintf("%d",GetFileUID(path))
	dat, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return "", err
	}
	userList := strings.Split(string(dat), "\n")
	for _, info := range userList[0 : len(userList)-1] {// 去掉最后一个\n 空元素
		// fmt.Println(info)
		s := strings.SplitN(info, ":", -1)
		if len(s) >= 3 && s[2] == uidStr {
			// fmt.Println(s[0])
			return s[0], nil
		}
	}
	return "", errors.New("error get fileOwner")
}

func StartFileMonitor(resultChan chan map[string]string) {
	log.Debug("%s","Start File Monitor and config the Watcher once...")
	var pathList []string
	Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Error("fsnotify Watcher error: %v",err)
		return
	}
	defer Watcher.Close()

	// 宿主主机监控
	for _,path := range common.Config.MonitorPath{
		if path == "%web%"{
			//TODO:web 目录的监控
		}

		// 找出要监控的目录:宿主主机一次遍历完成
		if strings.HasPrefix(path,"/"){
			pathList = append(pathList,path) // path 记录所有加入watch的目录
			if strings.HasSuffix(path,"*"){
				iterationWatcher([]string{strings.Replace(path,"*","",1)}, Watcher,pathList)
			}else {
				Watcher.Add(path) // 以文件夹为监控watcher
				log.Debug("add new wather: %v",strings.ToLower(path))
			}
		}else {
			log.Debug("error file monitor config! %v",path)
		}
	}
	// docker 主机监控
	// TODO: 识别由哪个docker的文件 监控事件
	go AddDockerWatch()

	var resultdata map[string]string
	for {
		select {
		case event := <-Watcher.Events:
			resultdata = make(map[string]string)
			//文件监控白名单过滤，监控自身文件夹变动不上报
			if common.InArray(pathList,strings.ToLower(event.Name),false) ||
				common.InArray(common.Config.Filter.File,strings.ToLower(event.Name),true){
				continue
			}
			if len(event.Name) == 0{
				continue
			}
			resultdata["source"] = "file"
			resultdata["action"] = event.Op.String()
			resultdata["path"] = event.Name
			resultdata["hash"] = ""
			resultdata["user"] = ""

			f, err := os.Stat(event.Name)
			if err == nil && !f.IsDir(){
				if f.Size() <= fileSize{
					if hash, err := getFileMD5(event.Name);err != nil{
						log.Error("get file hash failed: %v",err)
					}else {
						resultdata["hash"] = hash
						if common.InArray(common.Config.Filter.File,strings.ToLower(hash),false){
							continue
						}
					}
				}

				if user, err := getFileUser(event.Name);err == nil {
					resultdata["user"] = user
				}
			}

			if isFileWhite(resultdata){
				continue
			}
			resultChan <- resultdata
			log.Debug("[+]Watcher new event: %v",resultdata)
		case err := <- Watcher.Errors:
			log.Error("error: %v", err)
		}
	}
}

func AddDockerWatch() {
	ticker := time.NewTicker(time.Second * 30)
	status := 0
	go func() {
		docker := map[string]int{}
		docker_watch := map[string][]string{}
		for _ = range ticker.C{
			// 看是否有 新启动docker 与 容器的退出 => 缓存表
			// docker 的copy-on-write
			// /var/lib/docker/overlay2/f553c1fceb7ba14cc8cda6e7f4aba27493c11f3c34cfa05d44ce5c70d97233d8(包含init)/diff
			status = (status + 1)%100  // 当前存活的主机
			dirs, err := ioutil.ReadDir("/var/lib/docker/overlay2/")
			if err != nil {
				log.Error("open docker overlay2 error:%v",err)
			}
			// 遍历查找新启动docker 加入watch  与 更新存活 docker status
			for _, dir := range dirs {
				if dir.IsDir() {
					dirname := dir.Name()
					if ok := strings.Contains(dirname,"-init");ok{
						dockerlayer := strings.Split(dirname,"-")[0]
						if _,ok := docker[dockerlayer];ok {
							docker[dockerlayer] = status // 更新存活状态
							continue
						}else {
							// 新启动docker,加入watcher
							// TODO：在monitro path 前加上docker diff层
							docker[dockerlayer] = status
							for _,path := range common.Config.MonitorPath{
								if path == "%web%"{
									//TODO:web 目录的监控
								}
								// 找出要监控的目录:宿主主机
								if strings.HasPrefix(path,"/"){
									if strings.HasSuffix(path,"*"){
										docker_iterpath := fmt.Sprintf("/var/lib/docker/overlay2/%v/merged%v",
											dockerlayer,strings.Replace(path,"*","",1))
										paths := iterationWatcherDocker([]string{docker_iterpath}, Watcher)
										docker_watch[dockerlayer] = append(docker_watch[dockerlayer],paths...)
									}else {
										docker_path := fmt.Sprintf("/var/lib/docker/overlay2/%v/merged%v",dockerlayer,path)
										Watcher.Add(docker_path) // 以文件夹为监控watcher
										docker_watch[dockerlayer] = append(docker_watch[dockerlayer],docker_path)
										log.Debug("docker add new wather: [%v]",strings.ToLower(docker_path))
									}
								}
							}
						}
					}
				}
			}

			// 删除 已经 不存在的docker 容器
			for dockerlayer,s := range docker{
				if s != status {
					// 不存在的docker 容器，删除watch
					for _,path := range docker_watch[dockerlayer]{
						Watcher.Remove(path)
						log.Debug("remove docker watch:%v",path)
					}

					//var pathList []string
					//for _,path := range config.MonitorPath{
					//	if path == "%web%"{
					//		//TODO:web 目录的监控
					//	}
					//	// 找出要监控的目录:宿主主机
					//	if strings.HasPrefix(path,"/"){
					//		pathList = append(pathList,path)
					//		if strings.HasSuffix(path,"*"){
					//			docker_iterpath := fmt.Sprintf("/var/lib/docker/overlay2/%v/merged%v",
					//				dockerlayer,strings.Replace(path,"*","",1))
					//			iterationWatcherRemove([]string{docker_iterpath}, Watcher,pathList)
					//		}else {
					//			docker_path := fmt.Sprintf("/var/lib/docker/overlay2/%v/merged%v",dockerlayer,path)
					//			Watcher.Remove(docker_path) // 以文件夹为监控watcher
					//			log.Debug("remove wather: [%v]",strings.ToLower(docker_path))
					//		}
					//	}
					//}
					delete(docker_watch,dockerlayer)
					delete(docker,dockerlayer) // 防止重复删除
				}
			}
		}
	}()
}

func iterationWatcherDocker(monList []string, watcher *fsnotify.Watcher) []string {
	pathList := []string{}
	for _,p := range monList{
		filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				log.Error("docker file walk error: %v",err)
				return nil
			}
			if f.IsDir(){
				pathList = append(pathList,path)
				err = watcher.Add(path)
				log.Debug("docker add new wather: %v",path)
				if err != nil{
					log.Error("add docker file watcher error: %v %v",err,path)
				}
			}
			return nil
		})
	}
	return pathList
}
