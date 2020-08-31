package collect

import (
	"fmt"
	"testing"
)

func TestGetUser(t *testing.T) {
	var infodata Info
	infodata = GetUser()
	fmt.Printf("%+v",infodata)
}
