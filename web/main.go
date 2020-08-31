package main

import (
	"github.com/thonsun/puppy-hids/web/common/log"
	"github.com/thonsun/puppy-hids/web/routers"
	"github.com/thonsun/puppy-hids/web/setting"
	"fmt"
	"net/http"
)

func main() {
	log.SetLogLevel(log.DEBUG)
	router := routers.InitRouter()

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
		Handler:        router,
		ReadTimeout:    setting.ReadTimeout,
		WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}