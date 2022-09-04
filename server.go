package main

import (
	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/net"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

func NewTCPDefault() iface.IServer {
	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	return net.NewServer(log)
}

func NewWsDefault() iface.IServer {
	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	return net.NewWSServer(log)
}

func NewWSServerWithConfig(cnf config.GlobalObject) iface.IServer {
	config.GlobalObj = &cnf
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	return net.NewServer(log)
}
