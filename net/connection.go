package net

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

type Connection struct {
	//当前Conn属于哪个Server
	TCPServer iface.IServer
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//消息管理MsgID和对应处理方法的消息管理模块
	MsgHandler iface.IMsgHandle
	//告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc
	//有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	sync.RWMutex
	//链接属性
	property map[string]interface{}
	////保护当前property的锁
	propertyLock sync.Mutex
	//当前连接的关闭状态
	isClosed bool
}

func NewConnection(server iface.IServer, conn *net.TCPConn, connID uint32, msgHandler iface.IMsgHandle) *Connection {
	c := &Connection{
		TCPServer:   server,
		Conn:        conn,
		ConnID:      connID,
		MsgHandler:  msgHandler,
		msgBuffChan: make(chan []byte, config.GlobalObj.MaxMsgChanLen),
	}
	c.TCPServer.GetConnMgr().Add(c)
	return c
}

func (c *Connection) StartWriter() {
	c.TCPServer.Logger().Debug("conn writer is running", zap.String("address", c.RemoteAddr().String()))
	defer c.TCPServer.Logger().Debug("conn writer is exit", zap.String("address", c.RemoteAddr().String()))
	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					c.TCPServer.Logger().Warn("send buff to connection error,conn writer will be exit", zap.Error(err))
					return
				}
			} else {
				c.TCPServer.Logger().Warn("conection msg buff chan is closed")
				break
			}
		case <-c.ctx.Done():
			c.TCPServer.Logger().Info("writer is done")
			return
		}
	}
}

func (c *Connection) StartReader() {
	c.TCPServer.Logger().Debug("conn reader is running", zap.String("address", c.RemoteAddr().String()))
	defer c.TCPServer.Logger().Info("conn reader is exit", zap.String("address", c.RemoteAddr().String()))
	defer c.Stop()
	for {
		select {
		case <-c.ctx.Done():
			c.TCPServer.Logger().Info("conn reader is done")
			return
		default:
			headData := make([]byte, c.TCPServer.Packet().GetHeadLen())
			if _, err := io.ReadFull(c.Conn, headData); err != nil {
				c.TCPServer.Logger().Warn("read msg head err", zap.Error(err))
				return
			}
			msg, err := c.TCPServer.Packet().Unpack(headData)
			if err != nil {
				c.TCPServer.Logger().Error("unpacket err", zap.Error(err))
				return
			}
			var data []byte
			if msg.GetDataLen() > 0 {
				data = make([]byte, msg.GetDataLen())
				if _, err := io.ReadFull(c.Conn, data); err != nil {
					c.TCPServer.Logger().Error("read msg data err", zap.Error(err))
					return
				}
			}
			msg.SetData(data)
			req := Requset{
				conn: c,
				msg:  msg,
			}
			if config.GlobalObj.WorkerPoolSize > 0 {
				//已经启动工作池机制，将消息交给Worker处理
				c.MsgHandler.SendMsgToTaskQueue(&req)
			} else {
				//从绑定好的消息和对应的处理方法中执行对应的Handle方法
				go c.MsgHandler.DoMsgHandler(&req)
			}
		}
	}
}

func (c *Connection) Start() {
	c.TCPServer.Logger().Debug("starting conn ", zap.Uint32("connID", c.GetConnID()))
	c.ctx, c.cancel = context.WithCancel(context.Background())
	go c.StartReader()
	go c.StartWriter()
	c.TCPServer.CallOnConnStart(c)
	select {
	case <-c.ctx.Done():
		c.finalizer()
		return
	}

}
func (c *Connection) Stop() {
	c.cancel()
}
func (c *Connection) Context() context.Context {
	return c.ctx
}
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 同步调用，最好不用
func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	if c.isClosed == true {
		return errors.New("conn is closed where send msg")
	}
	dp := c.TCPServer.Packet()
	msg, err := dp.Pack(NewMessage(msgID, data))
	if err != nil {
		c.TCPServer.Logger().Warn("Pack error msg", zap.Error(err))
		return errors.New("Pack error msg ")
	}
	_, err = c.Conn.Write(msg)
	return err
}
func (c *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	c.RLock()
	defer c.RUnlock()
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	//将data封包，并且发送
	dp := c.TCPServer.Packet()
	msg, err := dp.Pack(NewMessage(msgID, data))
	if err != nil {
		c.TCPServer.Logger().Warn("Pack error msg", zap.Error(err))
		return errors.New("Pack error msg ")
	}

	// 发送超时
	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case c.msgBuffChan <- msg:
		return nil
	}
}
func (c *Connection) SetProperty(key string, value interface{}) {

}
func (c *Connection) GetProperty(key string) (interface{}, error) {

	return nil, nil
}
func (c *Connection) RemoveProperty(key string) {

}

func (c *Connection) finalizer() {
	//如果用户注册了该链接的关闭回调业务，那么在此刻应该显示调用
	c.TCPServer.CallOnConnStop(c)

	c.Lock()
	defer c.Unlock()

	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}

	c.TCPServer.Logger().Debug("Conn Stop()... ", zap.Uint32("connID", c.ConnID))

	// 关闭socket链接
	_ = c.Conn.Close()

	//将链接从连接管理器中删除
	c.TCPServer.GetConnMgr().Remove(c)

	//关闭该链接全部管道
	close(c.msgBuffChan)
	//设置标志位
	c.isClosed = true
}
