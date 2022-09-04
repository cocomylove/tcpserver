package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/net"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

func main() {
	config.InitGlobal()
	conf := zap.NewDevelopmentConfig()
	log, _ := conf.Build()
	//创建一个server句柄
	s := net.NewWSServer(log)

	//注册链接hook回调函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	// 多路由
	s.AddRouter(1, &Ping{})
	// s.AddRouter(2, &HelloRouter{})

	//2 开启服务
	s.Serve()
}

func DoConnectionBegin(conn iface.IConnection) {
	fmt.Println("DoConnectionBegin is Called ... ")
	err := conn.SendMsg(2, []byte("DoConnection BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}
}

//连接断开的时候执行
func DoConnectionLost(conn iface.IConnection) {
	fmt.Println("DoConnectionLost is Called ... ")
}

type Ping struct {
	net.BaseRouter
}

func (br *Ping) PreHandle(req iface.IRequest) {}

func (br *Ping) Handle(req iface.IRequest) {
	log.Println(string(req.GetData()))
	for i := 0; i < 100; i++ {
		req.GetConnection().SendMsg(req.GetMsgID(), []byte(strconv.Itoa(i)))
	}

}

func (br *Ping) PostHandle(req iface.IRequest) {}
