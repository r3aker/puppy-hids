[+]临时告警中心[+] level:0 info:IP:10.2.195.185,Type:process,Info:/bin/ping 33 0 uid != euid进程执行，可能为威胁进程 
//TODO
docker 容器 用户 执行 进程 映射到 宿主主机 的对应用户 执行 进程（进程pid不一样，namespace不同)
进程的path ppath 可能找不到
// docker 若没有启动namespace 隔离启动容器，默认都是root 的权限 这个误报

// ping 内网扫描的行为没有检测：ICMP 协议：只检测了TCP 的连接
// docker 里面的容器网络连接监控与文件变动？


