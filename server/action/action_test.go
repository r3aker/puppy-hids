package action

import (
	"testing"
	"time"
)


func TestComputerInfoSave(t *testing.T) {
	var info ComputerInfo = ComputerInfo{
		IP:       "0.0.0.0",
		System:   "Debain",
		Hostname: "localhost",
		Type:     "xxx",
		Path:     []string {"xxx"},
		Uptime:   time.Now(),
	}
	ComputerInfoSave(&info)
}

func TestDataInfoSave(t *testing.T) {
	var info DataInfo = DataInfo{
		IP:     "xxx",
		Type:   "xxx",
		System: "xxx",
		Data:   nil,
		Uptime: time.Time{},
	}
	DataInfoSave(&info)
}
