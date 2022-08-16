package net

type Message struct {
	DataLen uint32 //消息的长度
	ID      uint32 //消息的ID
	Data    []byte //消息的内容
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
