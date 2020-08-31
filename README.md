### PUPPY-HIDS USAGE
> 四个模块：
>
>web: 前端输入server检测规则rule、黑白名单,agent监控文件、audit 规则到mongodb；响应agent 定时获取在线server列表
>
>server：提供 agent rpc server调用，从mongodb中获取agent监控配置,保存server 在线信息到mongodb,发送日志到es进行可视化
>
>agent: netlink family 的NETLINK_AUDIT 协议监听SYSCALL时间，规则可以动态配置；读取/proc 文件系统需要管理员权限运行
>
>daemon: 响应server指令下发（socket）

系统采用netlink_audit协议检测进程execve,connect系统调用，仅依赖内核kauditd，这个在内核2.6默认集成；对于安装第三方auditd
用户程序服务需要停止，否则agent 收不到系统通过netlink套接字发送的事件信息,通过libpcap抓取网络连接五元组信息，在/proc补全进程argv,comm,path等信息，对与短暂进程、网络连接建立LRU缓存；支持docker 
容器检测；文件监控通过fsnotify 监控关键文件的创建,删除等行为，对于docker容器文件监控overlayer2的merge层。

### 单独模块启动顺序
[web]
go run main.go

[server]
go run server -db=192.168.8.114:27017 -es=192.168.8.114:9200

[agent]
sudo go run agent 192.168.8.243:8000 [debug]

### 正式生产
[web]
go run main.go

启动web api,写入server,agent 配置，如配置，告警规则到mongodb

[daemon]
sudo ./daemon -install -web 192.168.8.243:8000

daemon集成对agent 的管理：update,reload,uninstall

daemon支持响应server 下发指令：

    exec:执行shell命令
    
    kill:根据进程名kill -9 pid



### Ubuntu debug hids
ES,mongodb 安装与服务启动
kibana 对ES可视化

```shell script
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


