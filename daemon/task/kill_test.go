package task

import (
	"fmt"
	"testing"
)

func TestKillProcess(t *testing.T) {
	data := KillProcess("top")

	fmt.Print(data, len(data))
}
