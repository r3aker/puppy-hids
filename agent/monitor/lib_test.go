package monitor

import (
	"fmt"
	"testing"
)

func TestFindFileInfo(t *testing.T) {
	file := "/home/thonsun/workspace/shell/test.sh"
	user, err := getFileUser(file)
	if err != nil{
		panic(err)
	}
	fmt.Println(user)
}