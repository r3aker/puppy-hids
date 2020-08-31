package config

import "github.com/smallnest/rpcx/client"

const (
	AUTH_TOKEN           string          = "67080fc75bb8ee4a168026e5b21bf6fc"
	CONFIG_REFRESH_INTERVAL int = 60
	FAIL_MODE client.FailMode = client.Failtry
	SERVER_API string = "/json/serverlist"
	TESTMODE bool = true
	LOGFILE_PATH = "/var/log/puppy-hids"
	LOGFILE = "/var/log/puppy-hids/agent.log"
)


