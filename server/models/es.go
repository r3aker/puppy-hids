package models

import (
	"github.com/thonsun/puppy-hids/server/utils"
	"context"
	"encoding/json"
	"github.com/olivere/elastic"
	"strings"
	"time"
)


//var processMapping = `
//{
//	"properties": {
//		"datatype":{
//			"type":"keyword"
//		},
//		"data": {
//			"properties": {
//				"command": {
//					"type": "text",
//					"fields": {
//						"keyword": {
//							"ignore_above": 256,
//							"type": "keyword"
//						}
//					}
//				},
//				"name": {
//					"type": "text",
//					"fields": {
//						"keyword": {
//							"ignore_above": 128,
//							"type": "keyword"
//						}
//					}
//				},
//				"parentname": {
//					"type": "text",
//					"fields": {
//						"keyword": {
//							"ignore_above": 128,
//							"type": "keyword"
//						}
//					}
//				},
//				"pid": {
//					"type": "keyword"
//				},
//				"ppid": {
//					"type": "keyword"
//				}
//			}
//		},
//		"ip": {
//			"type": "ip"
//		},
//		"time": {
//			"type": "date"
//		}
//	}
//}`

// 集中所有的 process | connection | file 的ES map 到data 里面
var processMapping = `
{
	"properties": {
		"datatype":{
			"type":"keyword"
		},
		"ip": {
			"type": "ip"
		},
		"time": {
			"type": "date"
		}，
		"data": {
			"properties": {
				"command": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 256,
							"type": "keyword"
						}
					}
				},
				"name": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 128,
							"type": "keyword"
						}
					}
				},
				"parentname": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 128,
							"type": "keyword"
						}
					}
				},
				"pid": {
					"type": "keyword"
				},
				"ppid": {
					"type": "keyword"
				}
			}
		}
	}
}`


var fileMapping = `
{
	"properties": {
		"datatype":{
			"type":"keyword"
		},
		"data": {
			"properties": {
				"action": {
					"type": "keyword"
				},
				"path": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 256,
							"type": "keyword"
						}
					}
				},
				"hash":{
					"type" :"keyword"
				},
				"user":{
					"type" :"string",
					"fields": {
						"keyword": {
							"ignore_above": 40,
							"type": "keyword"
						}
					}
				}
			}
		},
		"ip": {
			"type": "ip"
		},
		"time": {
			"type": "date"
		}
	}
}`

var loginlogMapping = `
{
	"properties": {
		"datatype":{
			"type":"keyword"
		},
		"data": {
			"properties": {
				"username": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 40,
							"type": "keyword"
						}
					}
				},
				"hostname": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 40,
							"type": "keyword"
						}
					}
				},
				"remote": {
					"type": "text",
					"fields": {
						"keyword": {
							"ignore_above": 25,
							"type": "keyword"
						}
					}
				},
				"status": {
					"type": "keyword"
				}
			}
		},
		"ip": {
			"type": "ip"
		},
		"time": {
			"type": "date"
		}
	}
}`

var connectionMapping = `
{
	"properties": {
		"datatype":{
			"type":"keyword"
		},
		"data": {
			"properties": {
				"dir": {
					"type": "keyword"
				},
				"path": {
					"type": "text"
				},
				"remote": {
					"type": "text"
				},
				"local": {
					"type": "text"
				},
				"name":{
					"type":"text",
					"fields": {
						"keyword": {
							"ignore_above": 40,
							"type": "keyword"
						}
					}
				},
				"pid":{
					"type":"keyword"
				},
				"protocol": {
					"type": "keyword"
				},
				"ctype": {
					"type":"keyword"
				}
			}
		},
		"ip": {
			"type": "ip"
		},
		"time": {
			"type": "date"
		}
	}
}`

// 增加告警到 ES 可视化
//"raw" : "{\"action\":\"WRITE\",\"hash\":\"\",\"path\":\"/etc/sysop/juno/metrics/sedwwuUgV\",\"user\":\"\"}",
//"time" : ISODate("2020-05-27T10:36:32.627Z"),
//"type" : "file", data_type:alert_file
//"source" : "可疑文件操作",
//"level" : 0,
//"description" : "关键文件位置存在可疑文件操作",
//"ip" : "10.227.18.247",
//"info" : "/etc/sysop/juno/metrics/sedwwuUgV|WRITE",
//"status" : 0
//var alertMapping = `
//{
//	"properties": {
//		"datatype":{
//			"type":"keyword"
//		},
//		"data": {
//			"properties": {
//				"source": {
//					"type": "keyword"
//				},
//				"description": {
//					"type": "text",
//					"fields": {
//						"keyword": {
//							"ignore_above": 25,
//							"type": "keyword"
//						}
//					}
//				},
//				"raw": {
//					"type": "text"
//				},
//				"info":{
//					"type":"text",
//				},
//				"level":{
//					"type":"int"
//				},
//				"status": {
//					"type": "int"
//				}
//			}
//		},
//		"ip": {
//			"type": "ip"
//		},
//		"time": {
//			"type": "date"
//		}
//	}
//}`

var Client *elastic.Client

type ESSave struct {
	DataType string `json:"datatype"`
	IP string `json:"ip"`
	Data map[string]string `json:"data"`
	Time time.Time `json:"time"`
}



var esChan chan ESSave
var nowindicesName string

func init() {
	nowDate := time.Now().Local().Format("2006_01")
	nowindicesName = "monitor" + nowDate
	var err error
	// TODO：通过url 连接es
	Client, err = elastic.NewClient(elastic.SetURL("http://" + *es),elastic.SetSniff(false))
	if err != nil {
		utils.Error("Elastic NewClient error: %v", err.Error())
		panic(1)
	}
	indexNameList, err := Client.IndexNames()
	if err != nil {
		utils.Error("Client Get IndexNames error: %v", err.Error())
		return
	}
	// server 管理ES 中的索引 ｜ 不是ES 内部的ILM
	if !inArray(indexNameList, nowindicesName, false) {
		newIndex(nowindicesName)
	}
	esChan = make(chan ESSave, 2048) // 缓冲队列
}

//InsertThread ES异步写入线程
func InsertThread() {
	utils.Debug("Start Insert to ES thread...")
	var data ESSave
	p, err := Client.BulkProcessor().
		Name("puppy-hids").
		Workers(2).
		BulkActions(100).                // commit if # requests >= 100
		BulkSize(2 << 20).               	// commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). 			// commit every 30s
		Do(context.Background())
	if err != nil {
		utils.Error("start BulkProcessor: %v", err)
	}
	// bulkprocessor 插入数据
	for {
		data = <-esChan
		//TODO: 一个ES index 多个mapping
		p.Add(elastic.NewBulkIndexRequest().Index(nowindicesName).Doc(data))
	}
}

// InsertEs 将数据插入es
func InsertEs(data ESSave) {
	//TODO：channel 丢失ip:port 的数据
	esChan <- data
}

// QueryLogLastTime 查询ip最后一条登录日志的时间
func QueryLogLastTime(ip string) (string, error) {
	termQuery := elastic.NewTermQuery("ip", ip)
	searchResult, err := Client.Search("monitor*").Type("loginlog").Query(termQuery).Sort("time", false).Size(1).Do(context.Background())
	if err != nil {
		return "", err
	}
	if searchResult.Hits.TotalHits != 0 {
		var res map[string]interface{}
		result, err := searchResult.Hits.Hits[0].Source.MarshalJSON()
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(result, &res)
		if err != nil {
			return "", err
		}
		return res["time"].(string), nil
	}
	return "all", nil
}


func esCheckThread() {
	utils.Debug("check es health...")
	ticker := time.NewTicker(time.Second * 3600)// 小时级别 创建新的索引
	for _ = range ticker.C {
		nowDate := time.Now().Local().Format("2006_01")
		nowindicesName = "monitor" + nowDate
		indexNameList, err := Client.IndexNames()
		if err != nil {
			continue
		}
		if inArray(indexNameList, nowindicesName, false) {
			if time.Now().Local().Day() >= 28 {
				nextData := time.Now().Local().AddDate(0, 1, 0).Format("2006_01")
				indicesName := "monitor" + nextData
				if !inArray(indexNameList, indicesName, false) {
					newIndex(indicesName)
				}
			}
		} else {
			newIndex(nowindicesName)
		}
	}
}

func newIndex(name string) {
	utils.Debug("new es index name: %v", name)
	Client.CreateIndex(name).Do(context.Background())
	Client.PutMapping().Index(name).Type("process").BodyString(processMapping).Do(context.Background())
	Client.PutMapping().Index(name).Type("connection").BodyString(connectionMapping).Do(context.Background())
	Client.PutMapping().Index(name).Type("loginlog").BodyString(loginlogMapping).Do(context.Background())
	Client.PutMapping().Index(name).Type("file").BodyString(fileMapping).Do(context.Background())
	// 增加告警数据
	//Client.PutMapping().Index(name).Type("alert").BodyString(alertMapping).Do(context.Background())
}

func inArray(list []string, value string, like bool) bool {
	for _, v := range list {
		if like {
			if strings.Contains(value, v) {
				return true
			}
		} else {
			if value == v {
				return true
			}
		}
	}
	return false
}