package net

import (
	"github.com/cocomylove/tcpserver/iface"
	"github.com/cocomylove/tcpserver/utils/config"
	"go.uber.org/zap"
)

type MessageHandler struct {
	Apis       map[uint32]iface.IRouter
	WorkerPool uint32
	TaskQueue  []chan iface.IRequest
	logger     *zap.Logger
}

func NewMessageHandler(logger *zap.Logger) *MessageHandler {
	mh := &MessageHandler{
		Apis:       make(map[uint32]iface.IRouter),
		WorkerPool: config.GlobalObj.WorkerPoolSize,
		TaskQueue:  make([]chan iface.IRequest, config.GlobalObj.WorkerPoolSize, config.GlobalObj.WorkerPoolSize),
		logger:     logger,
	}
	return mh
}

func (m *MessageHandler) DoMsgHandler(request iface.IRequest) {
	handler, ok := m.Apis[request.GetMsgID()]
	if !ok {
		m.logger.Warn("api is not FOUND ", zap.Uint32("msgID", request.GetMsgID()))
		return
	}
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}
func (m *MessageHandler) AddRouter(msgID uint32, router iface.IRouter) {
	m.Apis[msgID] = router
}
func (m *MessageHandler) StartWorkerPool() {

}
func (m *MessageHandler) SendMsgToTaskQueue(request iface.IRequest) {
	id := request.GetConnection().GetConnID() % m.WorkerPool

	m.TaskQueue[id] <- request
}
