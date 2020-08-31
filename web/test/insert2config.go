package main

import "gopkg.in/mgo.v2/bson"

type client struct {//mongodb data
	TYPE string
	DIC ClientConfig
}

type ClientConfig struct {
	Cycle       int      `bson:"cycle"` // 信息传输频率，单位：分钟
	UDP         bool     `bson:"udp"`   // 是否记录UDP请求
	LAN         bool     `bson:"lan"`   // 是否本地网络请求
	Type        string   `bson:"type"`  // web | db | linux
	Filter      filter   `bson:",omitempty"`// 直接过滤不回传的数据
	MonitorPath []string `bson:"monitorPath"` // 监控目录列表
	Lasttime    string   `bson:",omitempty"`// 最后一条登录日志时间
}

type filterres struct {//mongodb data
	Type string `bson:"type"`
	Dic  filter `bson:"dic"`
}
type filter struct {
	File    []string `bson:"file"`    // 文件hash、文件名
	IP      []string `bson:"ip"`      // IP地址
	Process []string `bson:"process"` // 进程名、参数
	Port    []int `bson:"port"` // connections port白名单
}
var blacklist = blackList{
File:    []string{},
IP:      []string{},
Process: []string{},
Other:   []string{},
}
var whitelist = whiteList{
File:    []string{},
IP:      []string{"0.0.0.0","127.0.0.1"},
Process: []string{},
Other:   []string{},
}
var serverconfig = serverConfig{
Learn:        false,
OfflineCheck: false,
Private:      `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAyvWK2/MXVbGZ6KvFi/arvZLP6kJuhe2leC+754RM7rM1iOoT
pI4H8Kz+39RbN3IeLGoD9OP+Ug1ak6IGGqi1O67DgRVAYEKLT5dBuxTPvPmTpZ2R
E0rN0QxLX5Nx6pvP84g0APFVfmDKZcoqH6o/0YRzqKTM/brVltEQWRW/geAfGMgw
B05j0Kw51bCwa+T5xgb1/rzCIPkJlaKcoYMz52Gvt+ES8KaHxXxH0gwzb8MoxuvN
uPrGH2XsPOLU1fjSUhpfmLsXryOTpJZL8aIwEWy7KpMslWHMFy6KqMKd/tngYKb/
M3U9Dq87y0TB8ipn5k4q7OgRW74vs+uBASWuTQIDAQABAoIBAQCgZ7e7ZkDHBXqy
nS+gEvBg/6s7Zg2b49qnRnKn47Q544EoGHg10dsMqG030cnV5Gdcit1dquPRTaSM
kb1pTHUQGmuBlZ4vdehMmyrkBOq6XDYI6qNCMBjCY4kenJWn6kVMIBWZuSLsourT
0BfCtveBS8FtQ/KPhh5Q+CKHhdy5c1MQfNDO6b1Q+Da7bnHuV4YbUmQCm1tngPuL
wjGag0/C3IavIVxrymFRMferjRBnx9c9Sq5lfD1dNCdBPW5ZDqWpr4rllcmFsQ7C
rbONl4BOBoBzFAfyUmhDQhFHQFUoQo/+aLtKQfXRnlUmP3q2AWOsG5tCnab4RCjg
ENTygVMxAoGBAPfLWhl3Cm4ujpcRfRMOVpe5Lml7YxwFEz4ga1WAycLisIoaaDFr
YcGn/QefzRMvTyX77T8JvVnKwPnaFPFhEhOhwZ4vdfqYNQ3yiRuRlEU8CbP1X2s/
n3T+XJ2K1fp85IGE1NVRVe8k2FmoGd6sSJnILQJMEM3QfLds6Uay1kkzAoGBANGu
GvMO5g0vgJXg0pPVHf01tS8TZYVG4BsRRUm21RupPDNWTOo7wBzT9OqvIRN03KNM
hFyB/lLM9MuIAfvwcDYJs1aRvEpxuNM25xnmfn9RRU1DsOxNlSIvP09tCRDVA1i5
p7CFZTrgi/s+Awv4x3GRSuVhRrfwgGRE55e91yp/AoGBAIWwIrYmcWwsliWO++ny
DGnjMNUcCsatPkqAdyg0SaZpY1G/GYPAKYevuGYKozu8hHk7yC4AdTYim6axMCdi
dbw9wxYzCPXgdI9H0Q0cp+AKmjmLIqXcN42JRjKBGxz/kNEH90P3k+Nn/4mvlfV7
AdhmFVJt84r29rKHgfvwtIfdAoGAFMle0JO8iLgZ1kHoflFVXMHTSWxx1wmUs/o9
VTZz3/8iAbDfhSURQYpdsFpWPBiMuv+d65HThZ/d8MN19uT6KtFBXyapdPPbL800
kePAzJxg82zvgC2cyDvI2fXkPS/w2f3luuEujOyv0+Ns5+Xs17xgoWbIXPnRsJ8I
GonuZ78CgYEAziHm+nl/aKcKB3CKH6Uk3LM0ozgVFV+qlAiUdEghXzo8HZtVS93j
GYpMCXtpeTv9POh5hpdj5+JXvW1FNZTiS1VDvUS7ZymkKfJ7fhosVN4J8TTK7F9F
Z/txmst6IRGCp6eSWGSDylnlToZ+OOM890XJZahQUhzVX/pBj8Rmag0=
-----END RSA PRIVATE KEY-----`,
Cert:         `-----BEGIN CERTIFICATE-----
MIIC8DCCAdgCCQCGdN7w761dsDANBgkqhkiG9w0BAQsFADA6MQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxCzAJBgNVBAcMAkJKMREwDwYDVQQKDAhieXRlLmNvbTAe
Fw0yMDA1MTkxMzEwMzhaFw0zMDA1MTcxMzEwMzhaMDoxCzAJBgNVBAYTAkNOMQsw
CQYDVQQIDAJCSjELMAkGA1UEBwwCQkoxETAPBgNVBAoMCGJ5dGUuY29tMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyvWK2/MXVbGZ6KvFi/arvZLP6kJu
he2leC+754RM7rM1iOoTpI4H8Kz+39RbN3IeLGoD9OP+Ug1ak6IGGqi1O67DgRVA
YEKLT5dBuxTPvPmTpZ2RE0rN0QxLX5Nx6pvP84g0APFVfmDKZcoqH6o/0YRzqKTM
/brVltEQWRW/geAfGMgwB05j0Kw51bCwa+T5xgb1/rzCIPkJlaKcoYMz52Gvt+ES
8KaHxXxH0gwzb8MoxuvNuPrGH2XsPOLU1fjSUhpfmLsXryOTpJZL8aIwEWy7KpMs
lWHMFy6KqMKd/tngYKb/M3U9Dq87y0TB8ipn5k4q7OgRW74vs+uBASWuTQIDAQAB
MA0GCSqGSIb3DQEBCwUAA4IBAQC+64/bE/0BfHsAaMvMDWieTubaVQXo/fGmvyjV
9IfA3zyY+0boRM7GPN519E+OuyNrhw7YaO5ihqps1NrM/TRteREE2MD6FNq4AlhR
2OC65g3zZ9PUcwFQ/NpabLFZkDRzWYWMg43kC9pVoMQCa3+MtnDCyan8wwPuXted
S1u7xM2EI/m0V2bPB0LYS8gM/awH68/mcEaffk+tKcchUEuWy/6Qtg3g90Cml1bR
+FJzArSr9wam9qpYoP31IbDkt5N6RN8CwQedXZSWoLvpkwedhHWMZqXtgOOhHROr
Eg6Ow+vu8kH1sOPxxnwBG2OFJin8Xg1QWGQM1eNjJ+jBAnNO
-----END CERTIFICATE-----`,
Public:`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyvWK2/MXVbGZ6KvFi/ar
vZLP6kJuhe2leC+754RM7rM1iOoTpI4H8Kz+39RbN3IeLGoD9OP+Ug1ak6IGGqi1
O67DgRVAYEKLT5dBuxTPvPmTpZ2RE0rN0QxLX5Nx6pvP84g0APFVfmDKZcoqH6o/
0YRzqKTM/brVltEQWRW/geAfGMgwB05j0Kw51bCwa+T5xgb1/rzCIPkJlaKcoYMz
52Gvt+ES8KaHxXxH0gwzb8MoxuvNuPrGH2XsPOLU1fjSUhpfmLsXryOTpJZL8aIw
EWy7KpMslWHMFy6KqMKd/tngYKb/M3U9Dq87y0TB8ipn5k4q7OgRW74vs+uBASWu
TQIDAQAB
-----END PUBLIC KEY-----`,
}

var clientconfig = ClientConfig{
Cycle:       1,
UDP:         false,
LAN:         false,
Type:        "",
MonitorPath: []string{"/sbin/*","/etc/systemd/*","/bin/*","/usr/bin/*","/usr/sbin/*","/etc/","/etc/init.d/*"},
}
var filterdata = filter{
File:   []string{"\\.(png|js|css|jpg|gif|wolff|svg)$"},
IP:      []string{},
Process: []string{},
Port:	[]int{},
}

func insert2Config() {
	var err error

	if err != nil{
		panic(err)
	}
	err = DB.C("config").Insert(bson.M{"type":"server","dic":serverconfig})
	if err != nil{
		panic(err)
	}
	err = DB.C("config").Insert(bson.M{"type":"blacklist","dic":blacklist})
	if err != nil{
		panic(err)
	}
	err = DB.C("config").Insert(bson.M{"type":"whitelist","dic":whitelist})
	if err != nil{
		panic(err)
	}

	// client 监控配置：监控目录
	err  = DB.C("config").Insert(bson.M{"type":"client","dic":clientconfig})
	if err != nil{
		panic(err)
	}
	err  = DB.C("config").Insert(bson.M{"type":"filter","dic":filterdata})
	if err != nil{
		panic(err)
	}
}
