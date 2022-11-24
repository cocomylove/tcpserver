package config

import (
	"github.com/cocomylove/tcpserver/iface"
)

type GlobalObject struct {
	TCPServer iface.IServer //当前Zinx的全局Server对象
	Host      string        //当前服务器主机IP
	TCPPort   int           //当前服务器主机监听端口号
	Name      string        //当前服务器名称

	Version          string //当前Zinx版本号
	MaxPacketSize    uint32 //都需数据包的最大值
	MaxConn          int    //当前服务器主机允许的最大链接个数
	WorkerPoolSize   int //业务工作Worker池的数量 若为0，则表示不启动工作池
	MaxWorkerTaskLen uint32 //业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    int //SendBuffMsg发送消息的缓冲最大长度

	/*
		config file path
	*/
	ConfFilePath string

	/*
		logger
	*/
	LogDir  string //日志所在文件夹 默认"./log"
	LogFile string //日志文件名称   默认""  --如果没有设置日志文件，打印信息将打印至stderr
	LogLeve string //debug info warn  error -- 默认打开debug信息

}

var GlobalObj *GlobalObject

func InitGlobal() {
	//初始化GlobalObject变量，设置一些默认值
	GlobalObj = &GlobalObject{
		Name:             "tcpServer",
		Version:          "V1.0",
		TCPPort:          8999,
		Host:             "0.0.0.0",
		MaxConn:          50,
		MaxPacketSize:    2048,
		ConfFilePath:     "",
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
		LogDir:           "",
		LogFile:          "",
		LogLeve:          "debug",
	}

}
