package net

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsConnection struct {
	//当前Conn属于哪个Server
	WsServer iface.IServer
	//当前连接的socket TCP套接字
	Conn *websocket.Conn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler iface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	// 无缓冲通道
	msgChan chan []byte
	sync.RWMutex
	//链接属性
	property map[string]interface{}
	////保护当前property的锁
	propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool
	//消息类型 TextMessage 或 BinaryMessage之类
	messageType int
}

func NewWsConnection(s iface.IServer, conn *websocket.Conn, connID uint32, mh iface.IMsgHandle) *WsConnection {
	wsc := &WsConnection{
		WsServer:    s,
		Conn:        conn,
		MsgHandler:  mh,
		isClosed:    false,
		msgBuffChan: make(chan []byte, config.GlobalObj.MaxMsgChanLen),
		msgChan:     make(chan []byte, 1),
		ctx:         context.Background(),
	}
	s.GetConnMgr().Add(wsc)
	return wsc
}

func (ws *WsConnection) StartReader() {
	ws.WsServer.Logger().Debug("connection StartReader start ", zap.Uint32("cid", ws.ConnID))
	defer ws.WsServer.Logger().Debug("connection StartReader exit", zap.Uint32("cid", ws.ConnID))
	for {
		msgType, data, err := ws.Conn.ReadMessage()
		if err != nil {
			ws.WsServer.Logger().Warn("conn startReader read data err", zap.Error(err))
			goto readError
		}

		// 以客户端为准
		ws.messageType = msgType
		ws.WsServer.Logger().Debug("conn start recv from connid", zap.Uint32("connid", ws.ConnID))
		ws.WsServer.Logger().Debug("", zap.String("data", string(data)))
		msg := WsMessage{}
		json.Unmarshal(data, &msg)
		ws.WsServer.Logger().Debug("", zap.Any("message", msg))
		req := requestPool.Get().(*Requset)
		req.conn = ws
		req.msg = &msg
		select {
		case <-ws.ctx.Done():
			goto readClose
		default:
			if config.GlobalObj.WorkerPoolSize > 0 {
				ws.MsgHandler.SendMsgToTaskQueue(req)
			} else {
				ws.MsgHandler.DoMsgHandler(req)
			}
		}
	}
readError:
	ws.Stop()
readClose:
}
func (ws *WsConnection) StartWriter() {}

func (ws *WsConnection) Start() {
	go ws.StartReader()
	// 如果任务没启动就开启

	ws.WsServer.CallOnConnStart(ws)
}
func (ws *WsConnection) Stop() {}
func (ws *WsConnection) Context() context.Context {
	return ws.ctx
}
func (ws *WsConnection) GetTCPConnection() *net.TCPConn {
	return nil
}
func (ws *WsConnection) GetWSConnection() *websocket.Conn {
	return ws.Conn
}
func (ws *WsConnection) GetConnID() uint32 {
	return ws.ConnID
}
func (ws *WsConnection) RemoteAddr() net.Addr {
	return ws.RemoteAddr()
}
func (ws *WsConnection) SendMsg(msgID uint32, data []byte) error {
	ws.Conn.WriteMessage(int(msgID), data)
	return nil
}
func (ws *WsConnection) SendBuffMsg(msgID uint32, data []byte) error {
	return nil
}
func (ws *WsConnection) SetProperty(key string, value interface{}) {

}
func (ws *WsConnection) GetProperty(key string) (interface{}, error) {
	return nil, nil
}
func (ws *WsConnection) RemoveProperty(key string) {

}
