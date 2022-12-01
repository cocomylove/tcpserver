package tcpserver

import (
	"github.com/cocomylove/tcpserver/config"
	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/server"
	"go.uber.org/zap"
)

func NewTCPDefault() iface.IServer {
	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	return server.NewServer(log)
}

func NewWsDefault() iface.IServer {
	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	return server.NewWSServer(log, *config.GlobalObj)
}

func NewWSServerWithConfig(cnf config.GlobalObject) iface.IServer {
	config.GlobalObj = &cnf
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	return server.NewServer(log)
}
