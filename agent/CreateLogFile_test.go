package main

import (
	"puppy-hids/agent/common/log"
	"testing"
)

func TestCreateLogFile(t *testing.T) {
	CreateLogFile()
	log.Debug("debug log file")
}
