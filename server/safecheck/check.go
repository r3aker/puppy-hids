package safecheck

import (
	"github.com/thonsun/puppy-hids/server/action"
	"github.com/thonsun/puppy-hids/server/models"
	"github.com/thonsun/puppy-hids/server/utils"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var ScanChan = make(chan action.DataInfo,4096)

type stats struct {
	ServerList []string `bson:"server_list"`
	Count      int      `bson:"count"`
}

//mgo 序列化与反序列化需要大写等开头
type test struct{
	Uid string `bson:"uid"`
	Gid string `bson:"gid"`
	Name string `bson:"name"`
	Description string `bson:"description"`
}
type user struct {
	Data []test `bson:"data"`
}

// Check 安全基线检测引擎架构
type Check struct {
	Info action.DataInfo //带检测数据（agent 上报数据)
	V 	map[string]string  // 当前检测数据内容
	Value string //触发规则的信息
	Description string // 规则简介信息
	Source string // 警报来源
	Level int // 警报等级
	CStatistics *mgo.Collection // 统计表
	CNotice *mgo.Collection // 警报表
	CInfo *mgo.Collection // 用户，service 等信息表
}

// ScanMonitorThread 安全检测线程
func ScanMonitorThread() {
	utils.Debug("Start Scan Thread...")
	// 10个检测goroutine
	for i := 0; i < 10; i++ {
		go func() {
			c := new(Check)
			c.CStatistics = models.DB.C("statistics")
			// TODO：找合法用户，audit 规则下发
			c.CInfo = models.DB.C("info")
			c.CNotice = models.DB.C("notice")
			for {
				c.Info = <-ScanChan // agent client rpc putinfo 调用开始检测 ｜ 结合规则
				c.Run()
			}
		}()
	}
}

func (c *Check) Run() {
	for _,c.V = range c.Info.Data{
		c.BlackFilter()
		if c.WhiteFilter(){
			continue
		}
		c.auditCheck()
		c.socCheck()
		c.Rules()
	}
}

// audit 进程单独监控：rule 正则不好定义
func (c *Check) auditCheck()  {
	utils.Debug("audit 检测引擎...")
	switch c.Info.Type {
	case "process":
		ip := c.Info.IP // 根据IP去到mongodb 找用户
		users := user{}
		// 在不清楚mongodb 中数据结构情况下
		//res := bson.M{}
		//err := c.CInfo.Find(bson.M{"type":"userlist","ip":ip}).One(&res)
		//if err != nil {
		//	utils.Error("error:%#v",err)
		//}
		//utils.Info("test:%#v",res)
		err := c.CInfo.Find(bson.M{"type":"userlist","ip":ip}).Select(bson.M{"data":1}).One(&users)
		if err != nil {
			utils.Error("error:%v",err)
		}
		//utils.Info("find ip:%v users:%v",ip,users)
		// TODO:匹配是否是sbin 的运行,找出系统原本设置s位文件
		reg := regexp.MustCompile("^/sbin/.*|^/usr/sbin/.*")
		if reg.MatchString(strings.ToLower(c.V["name"])) {
			utils.Debug("正常s位系统文件执行:\n%v",c.Info)
			return
		}

		if c.V["uid"] != c.V["euid"]{
			c.Source = "uid != euid进程执行"
			c.Level = 0 // 根据规则生成告警的等级
			c.Description = "uid != euid进程执行，可能为威胁进程" // 进一步判断是否是正常的用户启动
			for _, user := range users.Data{
				if c.V["uid"] == user.Uid{
					utils.Debug("正常用户sudo")
					return
				}
			}
			c.Value = fmt.Sprintf("%s",c.V["name"]+" "+c.V["uid"]+" "+c.V["euid"])
			c.warning()
		}
	case "audit_config_change":
		c.Source = "audit 规则变化"
		c.Level = 0 // 根据规则生成告警的等级
		c.Description = "audit规则变化，audit监控被篡改"
		c.Value = fmt.Sprintf("%s",c.V["auid"]+c.V["op"])
		c.warning()
	default:
		//TODO:更多audit 监控规则: 如监控文件位置的执行:execue调用
		return
	}
}

func (c *Check) socCheck() {
	// TODO: 连接soc 威胁情报 恶意ip识别
}

// BlackFilter 黑名单检测 ：关键字是否在黑名单中
func (c *Check) BlackFilter() {
	var keyword string
	var blackList []string
	regex := true
	switch c.Info.Type {
	case "process":
		blackList = models.Config.BlackList.Process
		keyword = c.V["name"]
	case "connection":
		blackList = models.Config.BlackList.IP
		keyword = strings.Split(c.V["remote"], ":")[0]
		regex = false
	case "loginlog":
		blackList = models.Config.BlackList.IP
		keyword = c.V["remote"]
		regex = false
	case "file":
		blackList = models.Config.BlackList.File
		keyword = c.V["hash"]
		regex = false// 这个黑名单文件是 允许｜禁止 特定的文件
	case "crontab":
		blackList = models.Config.BlackList.File
		keyword = c.V["command"]
	default:
		blackList = models.Config.BlackList.Other
		keyword = c.V["name"]
	}
	// 检测到存在与黑名单中的 黑名单来源：自己配置
	if len(blackList) >= 1 && inArray(blackList, strings.ToLower(keyword), regex) {
		c.Source = "blacklist"
		c.Level = 0
		c.Description = "存在于黑名单列表中"
		c.Value = keyword
		c.warning()
	}
}

// WhiteFilter 白名单筛选
func (c *Check) WhiteFilter() bool {
	var keyword string
	var whiteList []string
	regex := true
	switch c.Info.Type {
	case "process":
		whiteList = models.Config.WhiteList.Process
		keyword = c.V["name"]
	case "connection":
		whiteList = models.Config.WhiteList.IP
		keyword = strings.Split(c.V["remote"], ":")[0]
		regex = false
	case "loginlog":
		whiteList = models.Config.WhiteList.IP
		keyword = c.V["remote"]
		regex = false
	case "file":
		whiteList = models.Config.WhiteList.File
		keyword = c.V["hash"]
		regex = false
	case "crontab":
		whiteList = models.Config.WhiteList.File
		keyword = c.V["command"]
	default:
		whiteList = models.Config.WhiteList.Other
		keyword = c.V["name"]
	}
	if len(whiteList) >= 1 && inArray(whiteList, strings.ToLower(keyword), regex) {
		return true
	}
	return false
}

// Rules 规则解析
func (c *Check) Rules() {
	utils.Debug("rule 检测引擎分析上报数据....")
	for _, r := range models.RuleDB {
		var vulInfo []string
		// 根据上报信息 找出符合的rule :os system && type(eg:linux and file)
		if (c.Info.System != r.System && r.System != "all") || c.Info.Type != r.Source {
			continue
		}
		i := len(r.Rules)
		if c.Info.Type == "connection"{
			utils.Debug("rule check: %v",c.Info)
		}
		//r.Rules map[string]rule (其中rule={TYPE:xxx,data:xxx})
		for k, rule := range r.Rules { // k是要检测的哪个上报数据字段,如process "name"
			switch rule.Type {
			case "regex":
				reg := regexp.MustCompile(rule.Data)
				if reg.MatchString(strings.ToLower(c.V[k])) {
					i--
					vulInfo = append(vulInfo, c.V[k])
				}

			case "non-regex":
				reg := regexp.MustCompile(rule.Data)
				if !reg.MatchString(strings.ToLower(c.V[k])) { // 正则表达式除外的匹配
					i--
					vulInfo = append(vulInfo, c.V[k])
				}
			case "string":
				if strings.ToLower(c.V[k]) == strings.ToLower(rule.Data) {
					i--
					vulInfo = append(vulInfo, c.V[k])
				}
			case "count":
				if models.Config.Learn {
					i--
					vulInfo = append(vulInfo, c.V[k])
					continue
				}
				var statsinfo stats
				var keyword string
				if c.Info.Type == "connection" {
					keyword = strings.Split(c.V[k], ":")[0] //remote ip

				} else {
					keyword = c.V[k]
				}
				err := c.CStatistics.Find(bson.M{"type": r.Source, "info": keyword}).One(&statsinfo) //如agent userlist 上报一次就不再上传了，为1
				if err != nil {
					utils.Error(err.Error(), r.Source, keyword)
					break
				}
				n, err := strconv.Atoi(rule.Data)
				if err != nil {
					utils.Error(err.Error())
					break
				}
				if statsinfo.Count == n {
					i--
					vulInfo = append(vulInfo, c.V[k])
				}
			}
		}

		if r.And { // 关联规则要全部匹配才算是符合规则
			if i == 0 {
				// 填充告警信息
				c.Source = r.Meta.Name
				c.Level = r.Meta.Level
				c.Description = r.Meta.Description
				sort.Strings(vulInfo)
				c.Value = strings.Join(vulInfo, "|")
				c.warning()
			}
		} else if i < len(r.Rules) {// 符合一条就告警
			c.Source = r.Meta.Name
			c.Level = r.Meta.Level // 根据规则生成告警的等级
			c.Description = r.Meta.Description
			sort.Strings(vulInfo)
			c.Value = strings.Join(vulInfo, "|")
			c.warning()
		}
	}
}

// 由warning 决定是否 产生告警事件
// status=4 是观察模式再次出现告警的事件
// status=3 是观察模式中的notice记录
// status=2 是忽略的notice记录
// status=0 将是输出告警信息的notice 记录
func (c *Check) warning() {
	// 观察模式 只记录统计 不显示
	utils.Debug("产生告警：%v",c.Info)
	if models.Config.Learn {
		// notice 库中 status=3是观察模式中的notice记录
		c.CNotice.Upsert(bson.M{"type": c.Info.Type, "ip": c.Info.IP, "source": c.Source, "level": c.Level,
			"info": c.Value, "description": c.Description, "status": 3, "time": c.Info.Uptime}, bson.M{"$inc": bson.M{"status": 1}})
	} else {
		// 如果忽略过就不写入 status=2 是忽略的notice记录 | ip + type + value 定位一条处理过的记录
		n, _ := c.CNotice.Find(bson.M{"type": c.Info.Type, "ip": c.Info.IP, "info": c.Value, "status": 2}).Count()
		if n >= 1 {
			return
		}
		raw, err := json.Marshal(c.V)
		if err != nil {
			utils.Debug("json 解析失败：%v",err)
		}
		err = c.CNotice.Insert(bson.M{"type": c.Info.Type, "ip": c.Info.IP, "source": c.Source, "level": c.Level,
			"info": c.Value, "description": c.Description, "status": 0, "raw": string(raw), "time": c.Info.Uptime})
		if err == nil {
			msg := fmt.Sprintf("IP:%s,Type:%s,Info:%s %s", c.Info.IP, c.Info.Type, c.Value, c.Description)
			sendNotice(c.Level, msg)
		}
	}
}
