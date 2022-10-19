package server

import (
	"errors"
	"sync"
	"time"

	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/ilog"
)

type ConnManager struct {
	connections map[uint32]iface.IConnection
	connLock    sync.RWMutex
	logger      ilog.Logger
}

func (cm *ConnManager) ConnGC() {
	currTime := time.Now().UnixMilli()
	for _, conn := range cm.connections {
		if conn.LastHeartbeatTime()+7000-currTime < 0 {

		}
	}
}

func NewConnManager(logger ilog.Logger) *ConnManager {
	cm := &ConnManager{
		connections: make(map[uint32]iface.IConnection),
		logger:      logger,
	}
	return cm
}

func (cm *ConnManager) Add(conn iface.IConnection) {
	cm.connLock.Lock()
	defer cm.connLock.Unlock()
	cm.connections[conn.GetConnID()] = conn
}
func (cm *ConnManager) Remove(conn iface.IConnection) {
	cm.connLock.Lock()
	delete(cm.connections, conn.GetConnID())
	cm.connLock.Unlock()
}

func (cm *ConnManager) Get(connID uint32) (iface.IConnection, error) {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	if conn, ok := cm.connections[connID]; ok {
		return conn, nil
	}
	return nil, errors.New("conn not found")
}
func (cm *ConnManager) Len() int {
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()
	return len(cm.connections)
}
func (cm *ConnManager) ClearConn() {
	cm.connLock.Lock()

	//停止并删除全部的连接信息
	for connID, conn := range cm.connections {
		//停止
		conn.Stop()
		//删除
		delete(cm.connections, connID)
	}
	cm.connLock.Unlock()
}
