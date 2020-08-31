package utils

import "testing"

func TestDebug(t *testing.T) {
	SetLogLevel(DEBUG)
	Debug("start xxx task...")
}
