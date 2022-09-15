package net

import (
	"fmt"
	"net"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/ilog"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

var startLogo = ""

// 服务器
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgID和对应的处理方法
	msgHandler iface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr iface.IConnManager
	//该Server的连接创建时Hook函数
	OnConnStart func(conn iface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn iface.IConnection)

	packet iface.IDataPack

	logger ilog.Logger
}

func NewServer(logger ilog.Logger, opt ...Option) *Server {
	printLogo()
	s := &Server{
		Name:       config.GlobalObj.Name,
		IPVersion:  "tcp4",
		IP:         config.GlobalObj.Host,
		Port:       config.GlobalObj.TCPPort,
		msgHandler: NewMessageHandler(logger),
		ConnMgr:    NewConnManager(logger),
		packet:     NewDataPack(),
		logger:     logger,
	}
	for _, o := range opt {
		o(s)
	}
	return s
}

func printLogo() {
	fmt.Println(startLogo)
}

func (s *Server) Start() {
	s.logger.Info("[START] server is starting", zap.String("host", s.IP), zap.Int("port", s.Port))
	go func() {
		s.msgHandler.StartWorkerPool()
		addres, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			s.logger.Error("", zap.Error(err))
			return
		}
		listener, err := net.ListenTCP(s.IPVersion, addres)
		if err != nil {
			s.logger.Error("", zap.Error(err))
			return
		}
		s.logger.Info("Server start succ....", zap.String("name", s.Name))

		//TODO: server.go 应该有一个自动生成ID的方法
		var connId uint32
		connId = 0
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				s.logger.Warn("Accpet err:", zap.Error(err))
				continue
			}
			s.logger.Debug("Get conn remote addr ", zap.String("RemoteAddress", conn.RemoteAddr().String()))
			if s.ConnMgr.Len() >= config.GlobalObj.MaxConn {
				conn.Close()
				continue
			}
			dealConn := NewConnection(s, conn, connId, s.msgHandler)
			connId++
			go dealConn.Start()
		}
	}()
}
func (s *Server) Stop() {
	s.logger.Warn("stop the server...")
	s.ConnMgr.ClearConn()

}
func (s *Server) Serve() {
	s.Start()
	//TODO: 启动后，需要加载的事情
	select {}
}
func (s *Server) AddRouter(msgID uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}
func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}
func (s *Server) SetOnConnStart(hook func(iface.IConnection)) {
	s.OnConnStart = hook
}
func (s *Server) SetOnConnStop(hook func(iface.IConnection)) {
	s.OnConnStop = hook
}
func (s *Server) CallOnConnStart(conn iface.IConnection) {
	if s.OnConnStart != nil {
		s.logger.Debug("start on conn...")
		s.OnConnStart(conn)
	}
}
func (s *Server) CallOnConnStop(conn iface.IConnection) {
	if s.OnConnStop != nil {
		s.logger.Debug("start on conn...")
		s.OnConnStop(conn)
	}
}
func (s *Server) Packet() iface.IDataPack {
	return s.packet
}

func (s *Server) Logger() ilog.Logger {
	return s.logger
}
func (s *Server) SetLogger(logger ilog.Logger) {
	s.logger = logger
}
