package server

import "sync"

type Message struct {
	DataLen uint32 `json:"data_len,omitempty"` //消息的长度
	ID      uint32 `json:"id,omitempty"`       //消息的ID
	Data    []byte `json:"data,omitempty"`     //消息的内容
}

var msgPool sync.Pool

func init() {
	msgPool = sync.Pool{
		New: func() any {
			return &Message{}
		},
	}
}

func NewMessage(id uint32, data []byte) *Message {
	return &Message{
		DataLen: uint32(len(data)),
		ID:      id,
		Data:    data,
	}
}

func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}
func (m *Message) GetMsgID() uint32 {
	return m.ID
}
func (m *Message) GetData() []byte {
	return m.Data
}
func (m *Message) SetMsgID(id uint32) {
	m.ID = id
}
func (m *Message) SetData(data []byte) {
	m.Data = data
}
func (m *Message) SetDataLen(length uint32) {
	m.DataLen = length
}

type WsMessage struct {
	DataLen uint32 `json:"data_len,omitempty"` //消息的长度
	ID      uint32 `json:"id,omitempty"`       //消息的ID
	Data    string `json:"data,omitempty"`     //消息的内容
}

func (m *WsMessage) GetDataLen() uint32 {
	return m.DataLen
}
func (m *WsMessage) GetMsgID() uint32 {
	return m.ID
}
func (m *WsMessage) GetData() []byte {
	return []byte(m.Data)
}
func (m *WsMessage) SetMsgID(id uint32) {
	m.ID = id
}
func (m *WsMessage) SetData(data []byte) {
	m.Data = string(data)
}
func (m *WsMessage) SetDataLen(length uint32) {
	m.DataLen = length
}
