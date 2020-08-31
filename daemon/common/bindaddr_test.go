package common

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestBindAddr(t *testing.T) {
	conn, _ := net.Dial("udp", fmt.Sprintf("%s:http","test.example.com"))
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	url := fmt.Sprintf("%s:65512", localAddr[0:idx])

	fmt.Printf(url)
}
