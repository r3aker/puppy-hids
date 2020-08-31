package setting

import "time"

const (
	HTTPPort = 8000
	ReadTimeout = 10 * time.Second
	WriteTimeout = 10 * time.Second

	RunMode = "debug"
)
