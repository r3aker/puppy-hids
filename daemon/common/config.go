package common

const (
	SERVER_LIST_API string = "/json/serverlist" // 仅接受server的tcp 连接请求
	Puplic_Key_API  string = "/json/publickey"  // server daemon 加密通讯
	TESTMODE        bool   = true               // API 接口使用http
	LOG_FILE               = "/var/log/puppy-hids/daemon.log"
	LOG_PATH               = "/var/log/puppy-hids"
)
