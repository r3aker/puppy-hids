package monitor

//type filterInfo struct {
//	Port    []int
//	Process []string
//	File    []string
//	IP		[]string
//}
//
//var filter filterInfo

const (
	fileSize int64 = 20480000
	UDP      uint8 = 17
	TCP      uint8 = 6
)

func init() {
	// 改从服务器获取配置
	// 硬编码白名单| 非正则匹配 去除上报数据
	//filter.Port = []int{137, 139, 445}
	//filter.File = []string{`c:\windows\temp`}
	//filter.IP = []string{"0.0.0.0","127.0.0.1"}
}
