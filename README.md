<h1 align="center">Welcome to puppy-hids 👋</h1>
<p>
  <img alt="Version" src="https://img.shields.io/badge/version-v1.0.0-blue.svg?cacheSeconds=2592000" />
  <a href="https://twitter.com/thonsun07" target="_blank">
    <img alt="Twitter: thonsun07" src="https://img.shields.io/twitter/follow/thonsun07.svg?style=social" />
  </a>
</p>

> a puppy version for  hids implementation,including process monitoring,filesystem monitoring and network monitoring.

## Theroy

通过netlink_audit协议检测进程execve,connect系统调用，仅依赖内核kauditd，这个在内核2.6默认集成；对于安装第三方auditd用户程序服务需要停止，否则agent 收不到系统通过netlink套接字发送的事件信息；

通过libpcap抓取网络连接五元组信息，在/proc补全进程argv,comm,path等信息，对与短暂进程、网络连接建立LRU缓存；

通过Linux系统提供的fsnotify实现监控关键文件的创建,删除等行为，对于docker容器文件监控overlayer2的merge层。

支持docker 容器检测：docker进程实际映射为宿主主机的一个进程，只是宿主主机和docker容器namespace不同，容器的流量经过宿主主机的网卡，所以可以实现对宿主主机上docker容器的进程监控、网络监控和文件监控。

系统由四个模块组成：

1. web: 前端输入server检测规则rule、黑白名单,agent监控文件、audit 规则到mongodb；响应agent 定时获取在线server列表

2. server：提供 agent rpc server调用，从mongodb中获取agent监控配置,保存server 在线信息到mongodb,发送日志到es进行可视化

3. agent: netlink family 的NETLINK_AUDIT 协议监听SYSCALL时间，规则可以动态配置；读取/proc 文件系统需要管理员权限运行

4. daemon: 响应server指令下发（socket），如agent的更新，停止，重启。

## Usage

### pre-order

1.系统依赖MongoDB，ES等，先确保MongoDB数据库配置了系统设置，可以在web/test/运行写入server启动需要的配置到MongoDB，可以按需求修改配置。

2.确保db保存了启动配置。server启动会注册地址到db，agent从启动参数获得web管理地址获取在线的server，将采集的数据通过rpc发送到server的流式事件引擎进行威胁发现。所以启动顺序是：db写入配置-server&web-agent。

3.生产环境可以安装deamon为系统服务，通过deamon进行agent的安装管理。

### develement

```sh
# 1.[web]
go run main.go

# 2.启动web api,写入server,agent 配置，如配置，告警规则到mongodb

# 3.[server]
go run server -db=192.168.8.114:27017 -es=192.168.8.114:9200

# 4.[agent]
sudo go run agent 192.168.8.243:8000 [debug]
```

### product

```sh
# 1.[web]
go run main.go

# 2.启动web api,写入server,agent 配置，如配置，告警规则到mongodb

# 3.[server]
go run server -db=192.168.8.114:27017 -es=192.168.8.114:9200

# 4.[daemon] 安装成系统服务
sudo ./daemon -install -web 192.168.8.243:8000

# daemon集成对agent 的管理：update,reload,uninstall
# daemon支持响应server 下发指令：
## exec:执行shell命令
## kill:根据进程名kill -9 pid
```



## Config

puppy-hids数据库详细

```sh
mongo --host 192.168.8.114 --port 27017 -u root -p root
> show dbs
admin   0.000GB
agent   0.000GB
config  0.000GB
local   0.000GB
> use agent
switched to db agent
> show collections
client
config
info
notice
queue
rules
server
statistics
task_result
```



## Author

👤 **thonsun**

* Website: http://thonsun.github.io
* Twitter: [@thonsun07](https://twitter.com/thonsun07)
* Github: [@thonsun](https://github.com/thonsun)

## Show your support

Give a ⭐️ if this project helped you!

***
_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_