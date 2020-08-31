package monitor

import (
	"github.com/thonsun/puppy-hids/agent/common"
	"github.com/thonsun/puppy-hids/agent/common/log"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/mdlayher/netlink"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	inetDiag = 4 // netlink_inet_addr 协议号
	SOCK_DIAG_BY_FAMILY netlink.HeaderType = 20 //netlink_inet_addr 协议依据socket id 获取socket 信息 消息类型
)

func StartNetSniff(resultChan chan map[string]string) {
	log.Debug( "%s","Start network monitor...")
	var resultData map[string]string

	// libpcap handler 获取
	h, err := getPcapHandle(common.LocalIP)
	if err != nil {
		log.Error("get dev handle error: %v", err)
		return
	}
	packetSource := gopacket.NewPacketSource(h, h.LinkType())
	packetSource.NoCopy = true

	// netlink 内核socket client建立
	//c, err := netlink.Dial(inetDiag,nil)
	//if err != nil{
	//	log.Fatalf("failed to dial netlink:%v",err)
	//}
	//defer c.Close()

	go func() {
		// 手动清除 不存在的PID 定时任务，减少资源占用
		ticker := time.NewTicker(time.Second*10)
		go func() {
			for _ = range ticker.C{
				for tpid,_ := range common.LocalPIDSocketCacheLRU.CacheMap{
					if _, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d",tpid));err != nil{
						log.Debug("pid %d not exist,clean the hash: %v",tpid,err)
						common.LocalPIDSocketCacheLRU.Remove(tpid) // 清除
					}
				}
			}
		}()
	}()


	for pkt := range packetSource.Packets() {
		if pkt == nil {
			continue
		}
		var socketid string
		var port int
		var target string // 目标主机

		var localPort int
		resultData = map[string]string{
			"source":   "",
			"dir":      "",
			"protocol": "",
			"remote":   "",
			"local":    "",
			"pid":      "",
			"name":     "",
			"cmdline":  "",
			"path":     "",
		}

		// 过滤IP
		// 不记录和server 的连接
		ipLayer := pkt.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, ok := ipLayer.(*layers.IPv4)
			srcIP := ip.SrcIP.String()
			dstIP := ip.DstIP.String()
			resultData["source"] = "connection"
			if common.LocalIP == srcIP {
				target = dstIP
				resultData["dir"] = "out" // 出去流量 记录目的IP
			} else {
				target = srcIP
				resultData["dir"] = "in" // 进来流量 记录来源IP
			}
			// TODO:识别服务器类型不记录web的tcp连接

			// 内网记录开关:不记录内网连接 LAN = true 记录内网
			if !common.Config.LAN && isLan(target) {
				//log.Debug("filter LAN:%s",target)
				continue
			}

			// 白名单过滤: 记录的目标IP 白名单过滤
			if common.InArray(common.Config.Filter.IP, target, false) {
				log.Debug("filter IP:%s",target)
				continue
			}

			//log.Info("new pcap: src:[%v] dst:[%v]",ip.SrcIP,ip.DstIP)
			if ok {
				if (common.LocalIP == srcIP && !common.InArray(common.ServerIPList, dstIP, false)) ||
					(common.LocalIP == dstIP && !common.InArray(common.ServerIPList, srcIP, false)) {
					switch ip.Protocol {
					case layers.IPProtocolTCP:
						//TODO 应该只记录 [syn]的握手包：
						//机器主动连接：syn -> syn|ack -> ack [记录这种情况]

						tcpLayer := pkt.Layer(layers.LayerTypeTCP)
						tcp, _ := tcpLayer.(*layers.TCP)
						//log.Info("new pcap: src:[%v:%v] dst:[%v:%v]",ip.SrcIP,tcp.SrcPort,ip.DstIP,tcp.DstPort)
						//找到tcp的主动发起握手信息
						if tcp.SYN && !tcp.ACK{
							log.Info("tcp new connection: [syn]:src:[%v:%v] dst:[%v:%v]",ip.SrcIP,tcp.SrcPort,ip.DstIP,tcp.DstPort)
							if tcpLayer != nil {
								tcp, _ := tcpLayer.(*layers.TCP)
								srcPort := splitPortService(tcp.SrcPort.String())
								dstPort := splitPortService(tcp.DstPort.String())
								log.Debug("port %v %v",srcPort, dstPort)
								// 记录TCP 的 syn连接请求
								resultData["protocol"] = "tcp"
								if resultData["dir"] == "out" {
									port, _ = strconv.Atoi(dstPort)
									localPort, _ = strconv.Atoi(srcPort)
								} else {
									port, _ = strconv.Atoi(srcPort)
									localPort, _ = strconv.Atoi(dstPort)
								}

								if isFilterPort(port) || isFilterPort(localPort) {
									log.Debug("filter port:%v %v",port,localPort)
									continue
								}
								resultData["remote"] = fmt.Sprintf("%s:%d", target, port)
								resultData["port"] = fmt.Sprintf("%d",port)
								resultData["local"] = fmt.Sprintf("%s:%d", common.LocalIP, localPort)
								// 增加LRU缓存
								//remote:local:tcp <=> socket inode 缓存 ｜
								//socket inode <=> pid 缓存 与 pid-socket inode缓存过期清理

								//定位流量 到 pid步骤
								// 1.libpcap 拿到src:sport -> dst:dport
								// 2.遍历/proc/net/tcp (docker里面的进程连接 tcp 记录不在这个 ) 找到socket inode
								// 3.拿到socket inode 遍历/proc/$pid/fd 打开的 socket inode, 定位到pid
								// TODO：docker 网络连接 为遍历查找 /proc/$pid/net/tcp
								// 改进1: 2,3 合并：通过流量遍历找 /proc/$pid/net/tcp 找到就直接定位socket inode,pid
								// 改进2：找出docker 进程映射｜ 文件映射

								// 宿主主机 遍历/proc/net/tcp
								resultData["ctype"] = "host"
								socketid, err = GetSocketID(port, target)
								if err != nil {
									log.Debug("[+]read /proc/net/tcp find socket inode error:%v",err)
									// 进入第二次docker 进程网络连接查找机会
									// TODO：标记进程 ｜ 网络为属于docker
									socketid, err = GetDockerSocketID(port,target)
									if err != nil {
										log.Error("no socket inode for connection:[%v:%v]",ip,port)
									}
									resultData["ctype"] = "docker"
								}
								if socketid == "0"{
									log.Debug("[+]TODO: socket inode 0,continue next pcap")
									// 重复连接为上一个connect调用
									continue
								}
								log.Debug("find socketid [%s] for [%v:%v]",socketid,target,port)
								//flag 标志是否LRU 定位 socket inode -> pid
								var flag bool = false //标志是否neng
								//短暂进程网络连接 快速查找 ：connect 时间生成一个缓存表
								// 方式一：遍历pid->[socket inode...] 配置
								//for pid_tmp,localSocketIDArry := range common.LocalPIDSocketInode {
								//	if common.InArray(localSocketIDArry,socketid,false) {
								//		resultData["pid"] = strconv.Itoa(pid_tmp)
								//		resultData["cmdline"] = common.LocalPIDInfo[pid_tmp]["cmdline"]
								//		resultData["name"] = common.LocalPIDInfo[pid_tmp]["name"]
								//		resultData["path"] = common.LocalPIDInfo[pid_tmp]["path"]
								//
								//		if _, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d",pid_tmp));err != nil{
								//			delete(common.LocalPIDSocketInode,pid_tmp) // 手动清除不存在pid
								//			delete(common.LocalPIDInfo,pid_tmp)
								//		}
								//		flag = true
								//		break
								//	}
								//}
								//全局hash 表过期配置｜ 查询时删除 or LRU
								//方式二：增加socket->pid 映射表O（1）查找
								//if pid,ok := common.LocalSocketInodePID[socketid];ok { // socketid -> pid存在
								//	resultData["pid"] = strconv.Itoa(pid)
								//	resultData["cmdline"] = common.LocalPIDInfo[pid]["cmdline"]
								//	resultData["name"] = common.LocalPIDInfo[pid]["name"]
								//	resultData["path"] = common.LocalPIDInfo[pid]["path"]
								//
								//	// 手动清除
								//	if _, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d",pid));err != nil{
								//		log.Debug("%s:%v","pid not exit,clean the hash",err)
								//		for _,v := range common.LocalPIDSocketInode[pid]{
								//			delete(common.LocalSocketInodePID,v)
								//		}
								//		delete(common.LocalPIDSocketInode,pid) // 手动清除不存在pid
								//		delete(common.LocalPIDInfo,pid)
								//	}
								//	flag = true
								//}
								// 方式三：LRU支持
								log.Debug("[+]PID INFO LRU Size: %d",common.LocalPIDSocketCacheLRU.Size())
								log.Debug("[+]PID INFO LRU DATA:\n%#v\n%#v",common.LocalSocketInodePID,common.LocalPIDInfo)
								if pid,ok := common.LocalSocketInodePID[socketid];ok{// socketid -> pid | LRU PID 要move to front
									resultData["pid"] = strconv.Itoa(pid)
									resultData["cmdline"] = common.LocalPIDInfo[pid]["cmdline"]
									resultData["name"] = common.LocalPIDInfo[pid]["name"]
									resultData["path"] = common.LocalPIDInfo[pid]["path"]
									common.LocalPIDSocketCacheLRU.Get(pid) // pid 前移

									flag = true
								}

								// 方式四：netlink_sock_addr 协议从内核获取socket inode 号
								//src_ip_num := ipToInt(srcIP)
								//src := [4]uint32{uint32(src_ip_num),0,0,0}
								//sport,_ := strconv.Atoi(srcPort)
								//dst_ip_num := ipToInt(dstIP)
								//dst := [4]uint32{uint32(dst_ip_num),0,0,0}
								//dport,_ := strconv.Atoi(dstPort)
								//
								//conn_req := inet_diag_req_v2{
								//	sdiag_family:syscall.AF_INET,
								//	sdiag_protocol:syscall.IPPROTO_TCP,
								//	idiag_stats: (1<<TCP_LISTEN) | (1<<TCP_ESTABLISHED) |
								//		(1<<TCP_TIME_WAIT) | (1<<TCP_SYN_SENT) | (1<<TCP_SYN_RECV) |
								//		(1 << TCP_FIN_WAIT1) | (1<< TCP_FIN_WAIT2),
								//	id:inet_diag_sockid{idiag_src:src,idiag_sport:uint16(sport),idiag_dst:dst,idiag_dport:uint16(dport)},
								//}
								//inet_data,err := conn_req.marshalBinary()
								//if err != nil {
								//	log.Error("data to byte error:%v",err)
								//}
								//req := netlink.Message{
								//	Header: netlink.Header{
								//		Type:SOCK_DIAG_BY_FAMILY,
								//		Flags:netlink.Request | netlink.Dump,
								//	},
								//	Data:   inet_data,
								//}
								//inet_req_status, err := c.Send(req)
								//if err != nil {
								//	log.Error("send req msg error:%v",err)
								//}
								//log.Debug("send req msg status:%v",inet_req_status)
								//inet_req_resp,err := c.Receive()
								//if err != nil {
								//	log.Error("recieve req resp error:%v",err)
								//}
								//for _,msg := range inet_req_resp{
								//	m := unmarshresp(msg.Data)
								//	fmt.Printf("[+]get inode: src:%v:%v dst:%v:%v user:%v inode:%v\n",IntToIP(uint(m.id.idiag_src[0])),
								//		unint16LE2BE(m.id.idiag_sport),IntToIP(uint(m.id.idiag_dst[0])),unint16LE2BE(m.id.idiag_dport),
								//		m.idiag_uid,m.idiag_inode)
								//}

								//TODO: 缓存表里面没有这个PID 信息：LRU过期 ｜ 重新加入
								if !flag {
									// 多数不要遍历/proc/$pid/fd 定位socket inode 对应的 pid
									pid, err := GetPID(socketid)
									if err != nil && pid == -1{
										log.Error("cannot find pid,next packet:%v",err)
										log.Error("socketid:%s",socketid)
										log.Error("ip:%s %s",resultData["remote"],resultData["local"])
										continue
									}
									resultData["pid"] = strconv.Itoa(pid)
									resultData["cmdline"] = findPIDCmdline(resultData["pid"])
									resultData["name"] = findPIDName(resultData["pid"])
									resultData["path"] = findPIDName(resultData["pid"])
								}
							}
							resultChan <- resultData
						}else {
							log.Debug("%s","not syn tcp packet,next")
						}
					// 当前配置不记录UDP 连接
					case layers.IPProtocolUDP:
						log.Debug("%s","udp new connection")
						if common.Config.UDP == false { // 不记录UDP
							log.Debug("%s","filter udp")
							continue
						}
						updLayer := pkt.Layer(layers.LayerTypeUDP)
						if updLayer != nil {
							udp, _ := updLayer.(*layers.UDP)
							srcPort := splitPortService(udp.SrcPort.String())
							dstPort := splitPortService(udp.DstPort.String())
							log.Debug("port: %v %v", srcPort, dstPort)
							resultData["protocol"] = "udp"
							if resultData["dir"] == "out" {
								port, _ = strconv.Atoi(dstPort)
								localPort, _ = strconv.Atoi(srcPort)
							} else {
								port, _ = strconv.Atoi(srcPort)
								localPort, _ = strconv.Atoi(dstPort)
							}

							// 从server 获取到client 的配置中
							if isFilterPort(port) || isFilterPort(localPort) {
								log.Debug("filter port:%v %v",port,localPort)
								continue
							}
							resultData["remote"] = fmt.Sprintf("%s:%d", target, port)
							resultData["local"] = fmt.Sprintf("%s:%d", common.LocalIP, localPort)
							socketid, err = GetSocketID(port, target)
							if err != nil {
								log.Error("find socket inode error:%v",err)
							}
							pid, err := GetPID(socketid)
							if err != nil && pid == -1{
								log.Error("cannot find pid,next packet:%v",err)
								log.Error("socketid:%s",socketid)
								log.Error("ip:%s:%s",resultData["remote"],resultData["local"])
								continue
							}
							resultData["pid"] = strconv.Itoa(pid)
							resultData["cmdline"] = findPIDCmdline(resultData["pid"])
							resultData["name"] = findPIDName(resultData["pid"])
							resultData["path"] = findPIDName(resultData["pid"])
							resultChan <- resultData
						}
					default:
						log.Debug("other type packet")
						continue
					}
				} else {
					continue
				}

			}
		}
	}
}


// 判断IP是否为内网
func isLan(ip string) bool {
	ipInt := ipToInt(ip)
	if (ipInt >= 167772160 && ipInt <= 184549375) || (ipInt >= 2886729728 && ipInt <= 2887778303) || (ipInt >= 3232235520 && ipInt <= 3232301055) {
		return true
	}
	return false
}

// 本地白名单过滤
// port 白名单过滤
func isFilterPort(port int) bool {
	for _,v := range common.Config.Filter.Port {
		if v == port {
			return true
		}
	}
	return false
}

// 获取网卡handle
func getPcapHandle(ip string) (*pcap.Handle, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return nil,err
	}
	var device string
	for _,dev := range devs{
		for _,v := range dev.Addresses{
			if v.IP.String() == ip {
				device = dev.Name
				log.Debug("libpcap listen on:%v %v",dev.Name,dev.Description)
				break
			}
		}
	}
	if device == ""{
		return nil, errors.New("find device error")
	}
	h, err := pcap.OpenLive(device,65535,true,0)// 混杂模式抓包
	if err != nil{
		return nil,err
	}
	log.Debug("Start network connection monitor: %v",device)
	err = h.SetBPFFilter("tcp or udp and (not broadcast and not multicast)")
	if err != nil {
		return nil, err
	}
	return h,nil
}
