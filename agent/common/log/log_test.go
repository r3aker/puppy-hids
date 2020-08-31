package log

import "testing"

func TestDebug(t *testing.T) {
	SetLogLevel(DEBUG)
	data := map[string]string{
		"ip":"xxxx",
		"port":"xxx",
		"proto":"xxx",
	}
	Debug("%#v",data)
}
