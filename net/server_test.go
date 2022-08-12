package net

import (
	"testing"

	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

func TestServer(t *testing.T) {

	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	s := NewServer(log)
	s.Serve()
}
