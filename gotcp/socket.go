package gotcp

import (
	"bytes"
	"sync"
	"syscall"
	"go_tcp/common/logging"
)


type SocketInfo struct {
	Fd             int
	Addr           syscall.Sockaddr
	LastAccessTime int64
	ReadBuffer     *bytes.Buffer
	WriteBuffer    *bytes.Buffer
	EpollFlag      int //EpollModFd比较耗时，可以保存上次epoll flag，暂时不优化
	Id			   uint64  //唯一标识
	WriteMutex sync.Mutex  //保护写缓冲区
}

func NewSocketInfo(fd int,id uint64, addr syscall.Sockaddr) *SocketInfo {
	return &SocketInfo{
		Fd:          fd,
		Id:			 id,
		Addr:        addr,
		ReadBuffer:  bytes.NewBuffer(make([]byte, 0)),
		WriteBuffer: bytes.NewBuffer(make([]byte, 0)),
		WriteMutex: sync.Mutex{},
	}
}

func (s *SocketInfo)AddMsgToWriteBuffer(msg []byte) bool{
	s.WriteMutex.Lock()
	nWrite, err :=s.WriteBuffer.Write(msg)
	s.WriteMutex.Unlock()
	if nWrite != len(msg) || err != nil {
		logging.Error("SocketInfo AddMsgToWriteBuffer Write Buffer failed,s=%+v",s )
		//s.closeConn(fd)
		return false
	}
	//logging.Debug("AddMsgToWriteBuffer msg:%#v",msg)
	return true
}


