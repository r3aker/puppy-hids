package monitor

import (
	"bytes"
	"encoding/binary"
	"github.com/mdlayher/netlink/nlenc"
)
/**
netlink_sock_addr | netlink_inet_diag协议 查询指定socket inode

 */

// socket identity
type inet_diag_sockid struct {
	idiag_sport uint16
	idiag_dport uint16
	idiag_src [4]uint32
	idiag_dst [4]uint32 // ipv4 32 | ipv6 128 的兼容
	idiag_if uint32
	idiag_cookie [2]uint32
}

// request structure
type inet_diag_req_v2 struct {
	sdiag_family uint8
	sdiag_protocol uint8
	sdiag_ext uint8
	pad uint8
	idiag_stats uint32
	id inet_diag_sockid
}
var order = nlenc.NativeEndian()

func (req inet_diag_req_v2) marshalBinary() ([]byte, error) {
	bytes := make([]byte,binary.Size(req))
	nlenc.PutUint8(bytes[0:1],req.sdiag_family)
	nlenc.PutUint8(bytes[1:2],req.sdiag_protocol)
	nlenc.PutUint8(bytes[2:3],req.sdiag_ext)
	nlenc.PutUint8(bytes[3:4],req.pad)
	nlenc.PutUint32(bytes[4:8],req.idiag_stats)
	//id 的写入
	nlenc.PutUint16(bytes[8:10],unint16LE2BE(req.id.idiag_sport))
	nlenc.PutUint16(bytes[10:12],unint16LE2BE(req.id.idiag_dport))
	nlenc.PutUint32(bytes[12:16],unint32LE2BE(req.id.idiag_src[0]))
	nlenc.PutUint32(bytes[28:32],unint32LE2BE(req.id.idiag_dst[0]))
	return bytes,nil
}

func (req inet_diag_req_v2) marshalBinary2()([]byte,error) {
	buf := bytes.NewBuffer(make([]byte,0,binary.Size(req)))
	binary.Write(buf,order,req.sdiag_family)
	binary.Write(buf,order,req.sdiag_protocol)
	binary.Write(buf,order,req.sdiag_ext)
	binary.Write(buf,order, req.pad)
	binary.Write(buf,order,req.idiag_stats)

	binary.Write(buf,order,req.id.idiag_sport)
	binary.Write(buf,order,req.id.idiag_dport)
	binary.Write(buf,order,req.id.idiag_src)
	binary.Write(buf,order,req.id.idiag_dst)
	return buf.Bytes(),nil
}

// resp struct
type inet_diag_msg struct {
	idiag_family uint8
	idiag_state uint8
	idiag_timer uint8
	idiag_retrans uint8
	id inet_diag_sockid
	idiag_expires uint32
	idiag_rqueue uint32
	idiag_wqueue uint32
	idiag_uid uint32
	idiag_inode uint32
}

func unmarshresp(data []byte) inet_diag_msg {
	resp := inet_diag_msg{}
	resp.idiag_family = nlenc.Uint8(data[0:1])
	resp.idiag_state = nlenc.Uint8(data[1:2])
	resp.idiag_timer = nlenc.Uint8(data[2:3])
	resp.idiag_retrans = nlenc.Uint8(data[3:4])

	resp.id.idiag_sport = nlenc.Uint16(data[4:6])
	resp.id.idiag_dport = nlenc.Uint16(data[6:8])
	resp.id.idiag_src[0] = nlenc.Uint32(data[8:12])
	resp.id.idiag_dst[0] = nlenc.Uint32(data[24:28])
	resp.id.idiag_if = nlenc.Uint32(data[40:44])
	resp.id.idiag_cookie[0] = nlenc.Uint32(data[44:48])
	resp.id.idiag_cookie[1] = nlenc.Uint32(data[48:52])

	resp.idiag_expires = nlenc.Uint32(data[52:56])
	resp.idiag_rqueue = nlenc.Uint32(data[56:60])
	resp.idiag_wqueue = nlenc.Uint32(data[60:64])
	resp.idiag_uid = nlenc.Uint32(data[64:68])
	resp.idiag_inode = nlenc.Uint32(data[68:72])
	return resp
}

func unmarshresp2(data []byte) inet_diag_msg {
	buf := bytes.NewBuffer(data)
	resp := inet_diag_msg{}
	binary.Read(buf,order,resp)
	return resp
}

// tcp socket state
type socket_stat uint32

const (
	_ socket_stat = iota
	TCP_ESTABLISHED
	TCP_SYN_SENT
	TCP_SYN_RECV
	TCP_FIN_WAIT1
	TCP_FIN_WAIT2
	TCP_TIME_WAIT
	TCP_CLOSE
	TCP_CLOSE_WAIT
	TCP_LAST_ACK
	TCP_LISTEN
	TCP_CLOSING
	TCPF_ALL = 0xFFF
)

const (
	INET_DIAG_NONE = iota
	INET_DIAG_MEMINFO
	INET_DIAG_INFO
	INET_DIAG_VEGASINFO
	INET_DIAG_CONG
	INET_DIAG_TOS
	INET_DIAG_TCLASS
	INET_DIAG_SKMEMINFO
	INET_DIAG_SHUTDOWN

	/*
	 * Next extenstions cannot be requested in struct inet_diag_req_v2:
	 * its field idiag_ext has only 8 bits.
	 */

	INET_DIAG_DCTCPINFO	/* request as INET_DIAG_VEGASINFO */
	INET_DIAG_PROTOCOL	/* response attribute only */
	INET_DIAG_SKV6ONLY
	INET_DIAG_LOCALS
	INET_DIAG_PEERS
	INET_DIAG_PAD
	INET_DIAG_MARK	/* only with CAP_NET_ADMIN */
	INET_DIAG_BBRINFO	/* request as INET_DIAG_VEGASINFO */
	INET_DIAG_CLASS_ID	/* request as INET_DIAG_TCLASS */
	INET_DIAG_MD5SIG
	INET_DIAG_ULP_INFO
	__INET_DIAG_MAX
)

func unint16LE2BE(s uint16) uint16 {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:],s)
	d := binary.BigEndian.Uint16(buf[:])
	return d
}

func unint32LE2BE(s uint32) uint32 {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:],s)
	d := binary.BigEndian.Uint32(buf[:])
	return d
}



