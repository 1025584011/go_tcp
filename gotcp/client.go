package gotcp

import (
	"go_tcp/common/logging"
)


type BusiInfo struct {
}

type ConnInfo struct {
	SInfo *SocketInfo
	BInfo *BusiInfo
	T     *IoThread
}

//同步发送消息，只能在io线程内使用
func (c *ConnInfo) SynSendMsg(msg []byte) bool {
	return c.T.WriteDirect(c.SInfo.Fd,msg) == nil
}

//异步发送消息
func (c *ConnInfo) AsynSendMsg(msg []byte) bool {
	if !c.SInfo.AddMsgToWriteBuffer(msg) {
		logging.Error("ConnInfo SendMsg failed,c:%+v",c)
		c.AsynClose()
		return false
	}
	c.T.Notify(EVENT_WRITE,c.SInfo)
	//logging.Debug("SendMsg msg:%#v,c:%+v",msg,c)
	return true
}

//同步关闭连接，只能在io线程内使用
func (c *ConnInfo) SynClose() bool {
	return c.T.CloseDirect(c.SInfo.Fd) == nil
}

//异步关闭连接
func (c *ConnInfo) AsynClose()  {
	logging.Debug("ConnInfo Close,c:%+v",c)
	c.T.Notify(EVENT_CLOSE,c.SInfo)
}