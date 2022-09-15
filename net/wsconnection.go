package net

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"

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
	ctx, cancel := context.WithCancel(context.Background())
	wsc := &WsConnection{
		WsServer:    s,
		Conn:        conn,
		MsgHandler:  mh,
		isClosed:    false,
		msgBuffChan: make(chan []byte, config.GlobalObj.MaxMsgChanLen),
		msgChan:     make(chan []byte, 1),
		ctx:         ctx,
		cancel:      cancel,
		ConnID:      connID,
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
func (ws *WsConnection) StartWriter() {
	ws.WsServer.Logger().Debug("conn writer is running", zap.String("address", ws.RemoteAddr().String()))
	defer ws.WsServer.Logger().Debug("conn writer is exit", zap.String("address", ws.RemoteAddr().String()))
	for {
		select {
		case data, ok := <-ws.msgBuffChan:
			if ok {
				if err := ws.Conn.WriteMessage(1, data); err != nil {
					ws.WsServer.Logger().Warn("send buff to connection error,conn writer will be exit", zap.Error(err))
					return
				}
			} else {
				ws.WsServer.Logger().Warn("conection msg buff chan is closed")
				break
			}
		case <-ws.ctx.Done():
			ws.WsServer.Logger().Info("writer is done")
			return
		}
	}
}

func (ws *WsConnection) Start() {
	go ws.StartReader()
	go ws.StartWriter()
	ws.WsServer.CallOnConnStart(ws)
	select {
	case <-ws.ctx.Done():
		ws.finalizer()
		return
	}
}
func (ws *WsConnection) Stop() {
	ws.cancel()
}
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
	return ws.Conn.RemoteAddr()
}
func (ws *WsConnection) SendMsg(msgID uint32, data []byte) error {
	ws.RLock()
	defer ws.RUnlock()
	if ws.isClosed == true {
		return errors.New("conn is closed where send msg")
	}
	if err := ws.Conn.WriteMessage(int(msgID), data); err != nil {
		return err
	}
	return nil
}

// ws发送的数据，全部使用[]byte 序列化方式由业务方面决定
// 通讯只关注数据本身
func (ws *WsConnection) SendBuffMsg(msgID uint32, data []byte) error {
	ws.RLock()
	defer ws.RUnlock()
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if ws.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	// 发送超时
	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case ws.msgBuffChan <- data:
		return nil
	}
}
func (ws *WsConnection) SetProperty(key string, value interface{}) {

}
func (ws *WsConnection) GetProperty(key string) (interface{}, error) {
	return nil, nil
}
func (ws *WsConnection) RemoveProperty(key string) {

}

func (c *WsConnection) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.WsServer.CallOnConnStop(c)

	c.Lock()
	defer c.Unlock()

	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}

	c.WsServer.Logger().Debug("Conn Stop()... ", zap.Uint32("connID", c.ConnID))

	// 关闭socket链接
	_ = c.Conn.Close()

	//将链接从连接管理器中删除
	c.WsServer.GetConnMgr().Remove(c)
	//关闭该链接全部管道
	close(c.msgBuffChan)
	//设置标志位
	c.isClosed = true
}
