package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/ilog"
	"github.com/cocomylove/tcpserver/utils/config"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsServer struct {
	cfg config.GlobalObject
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

	// connid 管理
	connId *ConnId
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	HandshakeTimeout: 5 * time.Second,
	// 解决跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWSServer(logger *zap.Logger, cfg config.GlobalObject) *WsServer {
	return &WsServer{
		cfg:        cfg,
		Name:       cfg.Name,
		Scheme:     "ws",
		IP:         cfg.Host,
		Port:       cfg.TCPPort,
		msgHandler: NewMessageHandler(logger),
		ConnMgr:    NewConnManager(logger),
		// packet:     NewDataPack(),
		Path:   "feature",
		logger: logger,
		connId: NewConnId(cfg.MaxConn),
	}
}

// ConnId 因为 connections 这个map删除key并非真正意义上的删除，为防止map内存泄露
// 固定key的值以及数量，可以方便map重用内存空间
type ConnId struct {
	bucket []uint32
	// maxConnId int
}

func NewConnId(maxConn int) *ConnId {
	c := &ConnId{
		bucket: make([]uint32, 0, maxConn),
	}

	for i := 0; i < maxConn; i++ {
		c.bucket = append(c.bucket, uint32(i))
	}
	return c
}

func (c *ConnId) Get() uint32 {
	id := c.bucket[0]
	c.bucket = c.bucket[1:]
	return id
}

func (c *ConnId) Put(id uint32) {
	c.bucket = append(c.bucket, id)
	log.Println("回收后bucket大小: ", len(c.bucket))
}

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("server wsHandler upgrade err", zap.Error(err))
		return
	}

	if s.ConnMgr.Len() >= s.cfg.MaxConn {
		s.logger.Warn("server wsHandler too many connection", zap.Int("maxConn", s.cfg.MaxConn))

		s.OnMaxConn(conn)
		_ = conn.Close()
		return
	}
	s.logger.Debug("server wsHandler a new conn ", zap.String("remoteAddr", conn.RemoteAddr().String()))
	id := s.connId.Get()
	s.logger.Debug("conid ", zap.Uint32("connid", id))
	dealConn := NewWsConnection(s, conn, id, int(s.cfg.MaxMsgChanLen), int(s.cfg.WorkerPoolSize), s.msgHandler)
	go dealConn.Start()

}

func (s *WsServer) Start() {
	s.OnMaxConn = func(conn *websocket.Conn) {
		_ = conn.Close()
	}
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
	// 回收id
	s.connId.Put(conn.GetConnID())

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
