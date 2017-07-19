package gotcp

import (
	"go_tcp/common/logging"
	"errors"
	"sync"
	"syscall"
	"time"
)

type IoThread struct {
	EventList        []syscall.EpollEvent
	NotifyList       []*NotifyEvent
	NotifyMutex      sync.Mutex
	NotifyWriteBytes [8]byte
	NotifyReadBytes  [8]byte
	Owner            *TcpServer
	EpollFd          int
	NotifyFd         int
	ReadTmpBuffer    []byte
}

func NewIoThread(o *TcpServer) *IoThread {
	if o == nil {
		logging.Error("NewIoThread invalid config")
		return nil
	}

	return &IoThread{
		Owner:            o,
		EventList:        make([]syscall.EpollEvent, int(o.MaxSocketNum/o.IoThreadNum)+1),
		NotifyList:       []*NotifyEvent{},
		NotifyMutex:      sync.Mutex{},
		ReadTmpBuffer:    make([]byte, READ_BUFFER_LEN),
		NotifyWriteBytes: [8]byte{1, 0, 0, 0, 0, 0, 0, 0},
		NotifyReadBytes:  [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (s *IoThread) Start() error {
	var err error
	s.EpollFd, err = syscall.EpollCreate(s.Owner.MaxSocketNum)
	if err != nil {
		logging.Error("IoThread EpollCreate failed")
		return err
	}
	//logging.Error("s.EpollFd=%d", s.EpollFd)
	r1, _, errn := syscall.Syscall(284, 0, 0, 0)
	if errn != 0 {
		logging.Error("SYS_EVENTFD failed,IoThread exit,errn=%d", errn)
		return nil
	}
	s.NotifyFd = int(r1)
	//logging.Error("s.NotifyFd=%d", s.NotifyFd)
	syscall.SetNonblock(s.NotifyFd, true)
	err = EpollAddFd(s.EpollFd, s.NotifyFd, syscall.EPOLLIN|syscall.EPOLLERR)
	if err != nil {
		logging.Error("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
		syscall.Close(s.NotifyFd)
		return err
	}

	go func() {
		lastNotify := time.Now()
		var ts int
		var b_handle_notify bool

		for {
			if time.Since(lastNotify) < time.Second {
				ts = 50
			} else {
				ts = 1000
			}
			nEvents, err := syscall.EpollWait(s.EpollFd, s.EventList, ts)
			if err != nil {
				logging.Error("IoThread EpollWait failed,err=%s", err.Error())
				continue
			}
			//time.Sleep(time.Microsecond * 100)
			//nEvents := 0
			//logging.Error("loop2,nEvents=%d,s.EpollFd=%d", nEvents, s.EpollFd)

			b_handle_notify = false
			for i := 0; i < nEvents; i++ {
				fd := int(s.EventList[i].Fd)
				if fd > s.Owner.MaxSocketNum {
					logging.Error("IoThread EpollWait invalid fd:%d", fd)
					continue
				}

				//有自定义事件
				if fd == s.NotifyFd {
					_, err := syscall.Read(s.NotifyFd, s.NotifyReadBytes[:])
					if err != nil {
						logging.Error("NotifyFd Read,err:%s", err.Error())
						continue
					}
					b_handle_notify = true
					continue
				}

				if s.EventList[i].Events&syscall.EPOLLERR > 0 {
					logging.Debug("IoThread EpollWait EPOLLERR")
					s.closeConn(fd)
					continue
				}

				if s.EventList[i].Events&syscall.EPOLLIN > 0 {
					s.handleRead(fd)
				}

				if s.EventList[i].Events&syscall.EPOLLOUT > 0 {
					s.handleWrite(fd)
				}
			}
			if b_handle_notify {
				lastNotify = time.Now()
				s.handleNotify()
			}
		}

	}()
	return nil
}

func (s *IoThread) handleNotify()  {
	logging.Debug("NotifyList len:%d", len(s.NotifyList))
	if len(s.NotifyList) != 0 {
		s.NotifyMutex.Lock()
		var nl = s.NotifyList
		s.NotifyList = []*NotifyEvent{} //置空
		s.NotifyMutex.Unlock()
		for _, v := range nl {
			if v.Type == EVENT_ACCEPT {
				s.handleAcceptEvent(v.Info)
			} else if v.Type == EVENT_WRITE {
				s.handleWriteEvent(v.Info)
			} else if v.Type == EVENT_CLOSE {
				s.handleCloseEvent(v.Info)
			}
		}
	}
}

func (s *IoThread) Notify(_type int, info *SocketInfo) error {
	e := &NotifyEvent{
		Type: _type,
		Info: info,
	}
	s.NotifyMutex.Lock()
	s.NotifyList = append(s.NotifyList, e)
	s.NotifyMutex.Unlock()
	_, err := syscall.Write(s.NotifyFd, s.NotifyWriteBytes[:])
	if err != nil {
		logging.Error("IoThread Notify failed,err:%s", err.Error())
		return err
	}
	return nil
}

func (s *IoThread) handleAcceptEvent(socketInfo *SocketInfo) error {
	logging.Debug("HandleAcceptEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	err := syscall.SetNonblock(socketInfo.Fd, true)
	if err != nil {
		logging.Error("IoThread HandleAccept SetNonblock failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	var flag = int(1)
	err = syscall.SetsockoptInt(socketInfo.Fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, flag)
	if err != nil {
		logging.Error("TcpServer Set TCP_NODELAY failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	err = EpollAddFd(s.EpollFd, socketInfo.Fd, syscall.EPOLLIN|syscall.EPOLLERR)
	if err != nil {
		logging.Error("IoThread HandleAccept EpollAddFd failed,err=%s", err.Error())
		syscall.Close(socketInfo.Fd)
		return err
	}

	connInfo := &ConnInfo{
		SInfo: socketInfo,
		T:     s,
	}
	s.Owner.ConnList[socketInfo.Fd] = connInfo
	return nil
}

func (s *IoThread) handleWriteEvent(socketInfo *SocketInfo) error {
	logging.Debug("HandleWriteEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !s.CheckSocketInfo(socketInfo) {
		logging.Error("HandleWriteEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}
	err := EpollModFd(s.EpollFd, socketInfo.Fd, syscall.EPOLLIN|syscall.EPOLLERR|syscall.EPOLLOUT)
	if err != nil {
		logging.Error("HandleWriteEvent EpollModFd failed,fd=%d,err=%s", socketInfo.Fd, err.Error())
		s.closeConn(socketInfo.Fd)
		return errors.New("sendMsg EpollModFd failed")
	}
	return nil
}

func (s *IoThread) handleCloseEvent(socketInfo *SocketInfo) error {
	logging.Debug("HandleCloseEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !s.CheckSocketInfo(socketInfo) {
		logging.Error("HandleCloseEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}

	s.closeConn(socketInfo.Fd)
	return nil
}

//异步场景下检查socket唯一id是否匹配
func (s *IoThread) CheckSocketInfo(socketInfo *SocketInfo) bool {
	return s.Owner.CheckSocketId(socketInfo.Fd, socketInfo.Id)
}

func (s *IoThread) closeConn(fd int) error {
	c := s.Owner.ConnList[fd]
	if c == nil {
		logging.Error("CloseConn already closed,fd=%d", fd)
		return nil
	}

	if c.SInfo.WriteBuffer.Len() != 0 {
		s.tryWrite(c.SInfo.Fd)
	}

	EpollDelFd(s.EpollFd, fd)
	syscall.Close(fd)
	s.Owner.ConnList[fd] = nil
	logging.Debug("CloseConn ok,SocketInfo:%+v", c.SInfo)
	return nil
}

func (s *IoThread) handleRead(fd int) error {
	connInfo := s.Owner.ConnList[fd]
	if connInfo == nil {
		logging.Error("IoThread HandleRead already closed,fd=%d", fd)

		return errors.New("HandleRead closed")
	}
	//connInfo.SInfo.LastAccessTime = time.Now().Unix()

	//tmp := make([]byte, ReadBufferLen)
	n, err := syscall.Read(fd, s.ReadTmpBuffer)
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		logging.Error("HandleRead Read error,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return errors.New("Read error")
	}

	//对端关闭，并且发送缓冲区无数据
	if n == 0 {
		logging.Info("HandleRead close by peer,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return nil
	}
	if n > READ_BUFFER_LEN {
		logging.Error("HandleRead read too much,n:%d", n)
		//panic("HandleRead read too much")
		return errors.New("read too much")
	}

	recv_msg := s.ReadTmpBuffer[0:n]
	//logging.Debug("HandleRead read msg:%#v", recv_msg)

	connInfo.SInfo.ReadBuffer.Write(recv_msg)

	var whole_msg []byte
	for {
		whole_msg = connInfo.SInfo.ReadBuffer.Bytes()
		ok, packlen := s.Owner.Parser.Unpack(whole_msg, connInfo)
		if !ok {
			logging.Error("HandleRead Unpack failed,whole_msg:%#v", whole_msg)
			s.closeConn(fd)
			return errors.New("Read error")
		}
		if packlen == 0 {
			logging.Debug("HandleRead incomplete pack,,whole_msg:%#v", whole_msg)
			break
		}

		if packlen > MAX_PACK_LEN {
			logging.Error("HandleRead invalid packlen:%d", packlen)
			s.closeConn(fd)
			return errors.New("invalid packlen")
		}

		if packlen > connInfo.SInfo.ReadBuffer.Len() {
			logging.Info("HandleRead need more,packlen:%d,current len:%d", packlen, connInfo.SInfo.ReadBuffer.Len())
			break
		}
		pack := connInfo.SInfo.ReadBuffer.Next(packlen)
		if !s.Owner.Parser.HandlePack(pack, connInfo) {
			logging.Error("HandleRead HandlePack failed,pack:%#v", pack)
			s.closeConn(fd)
			return errors.New("HandlePack failed")
		}
		//logging.Debug("HandleRead HandlePack ok,pack:%#v", pack)
		if 0 == connInfo.SInfo.ReadBuffer.Len() {
			logging.Debug("HandleRead HandlePack finished")
			break
		}
	}
	//测试
	//s.sendMsg(fd, s.ReadTmpBuffer[0:n])

	return nil
}

func (s *IoThread) tryWrite(fd int) error {
	return s.handleWrite(fd)
}

//在io线程中才可以调用
func (s *IoThread) CloseDirect(fd int) error {
	return s.closeConn(fd)
}

//在io线程中才可以调用
func (s *IoThread) WriteDirect(fd int,msg []byte) error {
	n, err := syscall.Write(fd, msg)
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		logging.Error("WriteDirect error,fd=%d,msg=%s", fd, string(msg))
		s.closeConn(fd)
		return errors.New("Write error")
	}
	return nil
}

func (s *IoThread) handleWrite(fd int) error {
	connInfo := s.Owner.ConnList[fd]
	if connInfo == nil {
		logging.Error("IoThread HandleWrite already closed,fd=%d", fd)
		return errors.New("HandleWrite closed")
	}
	//connInfo.SInfo.LastAccessTime = time.Now().Unix()

	writeBuffLen := connInfo.SInfo.WriteBuffer.Len()
	if writeBuffLen == 0 {
		logging.Debug("HandleWrite no data to write")
		return nil
	}

	connInfo.SInfo.WriteMutex.Lock()
	defer connInfo.SInfo.WriteMutex.Unlock()
	writeBuffLen = connInfo.SInfo.WriteBuffer.Len() //重新取一遍

	n, err := syscall.Write(fd, connInfo.SInfo.WriteBuffer.Bytes())
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		logging.Error("HandleWrite Write error,fd=%d,addr=%s", fd, connInfo.SInfo.Addr)
		s.closeConn(fd)
		return errors.New("Write error")
	}

	if n == writeBuffLen {
		//已发完，取消 EPOLLOUT事件
		err := EpollModFd(s.EpollFd, fd, syscall.EPOLLIN|syscall.EPOLLERR)
		if err != nil {
			logging.Error("HandleWrite EpollModFd failed,fd=%d,err=%s", fd, err.Error())
			s.closeConn(fd)
			return errors.New("HandleWrite EpollModFd failed")
		}
		connInfo.SInfo.WriteBuffer.Reset()
		//s.Owner.Parser.WriteFinishCb(connInfo)
		//logging.Error("HandleWrite Reset WriteBuffer,Cap=%d", connInfo.SInfo.WriteBuffer.Cap())
	} else {
		//修改WriteBuffer的偏移
		connInfo.SInfo.WriteBuffer.Next(n)
	}

	return nil
}

/*
func (s *IoThread) sendMsg(fd int, msg []byte) error {
	connInfo := s.Owner.ConnList[fd]
	if connInfo == nil {
		logging.Error("IoThread HandleSendMsg already closed,fd=%d", fd)
		return errors.New("HandleSendMsg closed")
	}
	nWrite, err := connInfo.SInfo.WriteBuffer.Write(msg)
	if nWrite != len(msg) || err != nil {
		logging.Error("HandleSendMsg Write Buffer failed,fd=%d", fd)
		s.closeConn(fd)
		return errors.New("HandleSendMsg Write Buffer failed")
	}

	err = EpollModFd(s.EpollFd, fd, syscall.EPOLLIN|syscall.EPOLLERR|syscall.EPOLLOUT)
	if err != nil {
		logging.Error("sendMsg EpollModFd failed,fd=%d,err=%s", fd, err.Error())
		s.closeConn(fd)
		return errors.New("sendMsg EpollModFd failed")
	}

	//logging.Debug("sendMsg msg:%s", string(msg))
	return nil
}
*/
