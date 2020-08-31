package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/willf/bitset"
)

type Notice struct {
	Type string `bson:"type"`
	Source string `bson:"source"`
	Description string `bson:"description"`
	IP string `bson:"ip"`
	Raw string `bson:"raw"`
	Info string `bson:"info"`
	Time time.Time `bson:"time"`
}

type ProcessInfo struct {
	Auid        string `json:"auid"`
	Cmdline     string `json:"cmdline"`
	Comm        string `json:"comm"`
	Epoch       string `json:"epoch"`
	Euid        string `json:"euid"`
	Exit        string `json:"exit"`
	Gid         string `json:"gid"`
	Key         string `json:"key"`
	Logid       string `json:"logid"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Pcmdline    string `json:"pcmdline"`
	Pid         string `json:"pid"`
	Pname       string `json:"pname"`
	Ppath       string `json:"ppath"`
	Ppid        string `json:"ppid"`
	ProcessType string `json:"process_type"`
	Success     string `json:"success"`
	Suid        string `json:"suid"`
	SyscallID   string `json:"syscall_id"`
	UID         string `json:"uid"`
}

type ConnectionInfo struct {
	Cmdline  string `json:"cmdline"`
	Ctype    string `json:"ctype"`
	Dir      string `json:"dir"`
	Local    string `json:"local"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Pid      string `json:"pid"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Remote   string `json:"remote"`
}

type FileInfo struct {
	Action string `json:"action"`
	Hash   string `json:"hash"`
	Path   string `json:"path"`
	User   string `json:"user"`
}

// 告警统计
/*
{
	"time":{
		"ip":{
			"type":{
				"success":xxxx
				"failed":xxxx
			}
		}
	}
}
 */

func getAlertFromDB() {
	var notices []Notice
	var fileInfo FileInfo
	var connectionInfo ConnectionInfo
	var processInfo ProcessInfo
	var err error
	noticeSum := make(map[string]map[string]map[string]map[string]int)
	//filter := NewBloomFilter()

	err = DB.C("notice").Find(nil).Sort("-time").All(&notices)
	if err != nil {
		fmt.Printf("find mongodb notice error:%v",err)
	}
	for _,notice := range notices {
		cst, _ := time.LoadLocation("Asia/Shanghai")
		notice.Time = notice.Time.In(cst)
		year,mon,day := notice.Time.Date()
		//map 指针类型需要初始化
		mapTimeIndex := fmt.Sprintf("%d-%d-%d",year,mon,day)
		//time.Now().Sub(notice.Time).Hours()/24<1
		if _,ok := noticeSum[mapTimeIndex];!ok {
			noticeSum[mapTimeIndex] = make(map[string]map[string]map[string]int)
		}
		if _,ok := noticeSum[mapTimeIndex][notice.IP];!ok {
			noticeSum[mapTimeIndex][notice.IP] = make(map[string]map[string]int)
		}
		if _,ok := noticeSum[mapTimeIndex][notice.IP][notice.Type];!ok {
			noticeSum[mapTimeIndex][notice.IP][notice.Type] = make(map[string]int)
		}

		switch notice.Type {
		case "file":
			err = json.Unmarshal([]byte(notice.Raw),&fileInfo)
			if err != nil {
				panic(1)
			}
			noticeSum[mapTimeIndex][notice.IP][notice.Type]["success"] += 1
		case "process":
			err = json.Unmarshal([]byte(notice.Raw),&processInfo)
			if err != nil {
				panic(1)
			}
			//if !filter.contains(processInfo.Logid) {
			//	continue
			//}else {
			//	filter.add(processInfo.Logid)
			//}
			if processInfo.Success == "yes" {
				noticeSum[mapTimeIndex][notice.IP][notice.Type]["success"] += 1
			}else {
				noticeSum[mapTimeIndex][notice.IP][notice.Type]["failed"] += 1
			}
		case "connection":
			err = json.Unmarshal([]byte(notice.Raw),&connectionInfo)
			if err != nil {
				panic(1)
			}
			noticeSum[mapTimeIndex][notice.IP][notice.Type]["success"] += 1
		default:
			fmt.Println(notice.Type)
		}

		fmt.Printf("[+]临时告警[+]:\nip:%v type:%v source:%v\ndecription:%v\ninfo:%v\ndata:%s\ntime:%v\n\n",notice.IP,
			notice.Type,notice.Source,notice.Description,notice.Info,notice.Raw,notice.Time.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("%+v\n",noticeSum)
	for day,map1 := range noticeSum{
		for ip,map2 := range map1{
			for atype,v := range map2{
				fmt.Printf("%s %s %s: %v\n",day,ip,atype,v)
			}
		}
	}
}



const DEFAULT_SIZE = 2<<24
var seeds = []uint{7, 11, 13, 31, 37, 61}

type BloomFilter struct {
	set *bitset.BitSet
	funcs [6]SimpleHash
}

func NewBloomFilter() *BloomFilter {
	bf := new(BloomFilter)
	for i:=0;i< len(bf.funcs);i++{
		bf.funcs[i] = SimpleHash{DEFAULT_SIZE,seeds[i]}
	}
	bf.set = bitset.New(DEFAULT_SIZE)
	return bf
}

func (bf BloomFilter) add(value string){
	for _,f:=range(bf.funcs){
		bf.set.Set(f.hash(value))
	}
}

func (bf BloomFilter) contains(value string) bool  {
	if(value == ""){
		return false
	}
	ret := true
	for _,f:=range(bf.funcs){
		ret = ret && bf.set.Test(f.hash(value))
	}
	return ret
}


type SimpleHash struct{
	cap uint
	seed uint
}

func (s SimpleHash) hash(value string) uint {
	var result uint = 0
	for i := 0; i < len(value); i++ {
		result = result*s.seed + uint(value[i])
	}
	return (s.cap - 1) & result
}