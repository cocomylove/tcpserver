package net

import (
	"net/http"
	"strconv"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/ilog"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type WsServer struct {

	//服务器的名称
	Name string
	// wss or ws
	Scheme string
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

	// 超过最大连接回调
	OnMaxConn func(conn *websocket.Conn)
	// 路径
	Path string

	// 任务
	logger ilog.Logger
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,

	// 解决跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSServer(logger *zap.Logger) *WsServer {
	return &WsServer{
		Name:       config.GlobalObj.Name,
		Scheme:     "ws",
		IP:         config.GlobalObj.Host,
		Port:       config.GlobalObj.TCPPort,
		msgHandler: NewMessageHandler(logger),
		ConnMgr:    NewConnManager(logger),
		// packet:     NewDataPack(),
		Path:   "feature",
		logger: logger,
	}
}

var cid uint32

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("server wsHandler upgrade err", zap.Error(err))
		return
	}

	if s.ConnMgr.Len() >= config.GlobalObj.MaxConn {
		s.logger.Warn("server wsHandler too many connection", zap.Int("maxConn", config.GlobalObj.MaxConn))
		s.OnMaxConn(conn)
		conn.Close()
		return
	}
	s.logger.Debug("server wsHandler a new conn ", zap.String("remoteAddr", conn.RemoteAddr().String()))
	dealConn := NewWsConnection(s, conn, cid, s.msgHandler)
	go dealConn.Start()
	cid++

}

func (s *WsServer) Start() {
	s.logger.Info("server start ", zap.String("name", s.Name))
	go s.msgHandler.StartWorkerPool()
	http.HandleFunc("/"+s.Path, s.wsHandler)
	err := http.ListenAndServe(s.IP+":"+strconv.Itoa(s.Port), nil)
	if err != nil {
		s.logger.Error("wsserver start faild", zap.Error(err))
	}
}
func (s *WsServer) Stop() {
	s.logger.Warn("ws server will stop ")
	s.ConnMgr.ClearConn()
}
func (s *WsServer) Serve() {
	//TODO: 启动前工作
	s.Start()
}
func (s *WsServer) AddRouter(msgID uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}
func (s *WsServer) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}
func (s *WsServer) SetOnConnStart(fn func(iface.IConnection)) {
	s.OnConnStart = fn
}
func (s *WsServer) SetOnConnStop(fn func(iface.IConnection)) {
	s.OnConnStop = fn
}
func (s *WsServer) CallOnConnStart(conn iface.IConnection) {
	if s.OnConnStart != nil {
		s.OnConnStart(conn)
	}

}
func (s *WsServer) CallOnConnStop(conn iface.IConnection) {
	if s.OnConnStop != nil {
		s.OnConnStop(conn)

	}

}
func (s *WsServer) Packet() iface.IDataPack {
	return nil
}
func (s *WsServer) Logger() ilog.Logger {
	return s.logger
}

func (s *WsServer) SetLogger(logger ilog.Logger) {
	s.logger = logger
}
