package net

type Messaage struct {
	DataLen uint32 //消息的长度
	ID      uint32 //消息的ID
	Data    []byte //消息的内容
}

func (m *Messaage) GetDataLen() uint32 {
	return m.DataLen
}
func (m *Messaage) GetMsgID() uint32 {
	return m.ID
}
func (m *Messaage) GetData() []byte {
	return m.Data
}
func (m *Messaage) SetMsgID(id uint32) {
	m.ID = id
}
func (m *Messaage) SetData(data []byte) {
	m.Data = data
}
func (m *Messaage) SetDataLen(length uint32) {
	m.DataLen = length
}
