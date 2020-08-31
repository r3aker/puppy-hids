package common

import (
	"fmt"
	"testing"
)

func TestFileMD5String(t *testing.T) {
	file := "../upload/agent"
	md5str,err := FileMD5String(file)
	if err != nil{
		fmt.Printf("error:%v",err)
	}
	fmt.Printf("%s",md5str)
}