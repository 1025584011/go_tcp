package gotcp

import (
	"errors"
	"go_tcp/common/logging"
	"strings"
	"syscall"
)

var(
	Server *TcpServer
)

/*
下面接口中的两个方法都是在io线程中调用的，不能有任何阻塞操作，业务中有阻塞场景请起协程，然后使用ConnInfo中的异步接口。
Unpack 用来计算包长；遇到非法包才返回false；未判断出包长时packlen返回0；成功时packlen返回包长
HandlePack 用来处理业务包，参数中的msg是完整包
*/
type TcpParser interface{
	Unpack(msg []byte,c *ConnInfo)(ok bool,packlen int)   //返回成功失败，包长,包长为0表示包长未知
	HandlePack(msg []byte,c *ConnInfo)(ok bool)
	//WriteFinishCb(c *ConnInfo)
}


//io线程数量不要超过 CPU物理core的个数（非逻辑处理器个数），配置为core-1 时性能最强
//查看core个数：cat /proc/cpuinfo| grep "cpu cores"| uniq
func InitServer(ioNum, maxSocketNum,checkTimeoutTs,timeoutTs int,addr string,parser TcpParser){
	Server = NewTcpServer(ioNum, maxSocketNum,checkTimeoutTs,timeoutTs,addr,parser)
	Server.Start()
}


type TcpServer struct {
	ConnList     []*ConnInfo
	MaxSocketNum int
	IoThreadNum  int
	IoThreadList []*IoThread
	Addr string
	UniqueId uint64
	Parser TcpParser
	CheckTimeoutTs int  //多久检查一次
	TimeoutTs int //多少秒超时
}

func NewTcpServer(ioNum, maxSocketNum,checkTimeoutTs,timeoutTs int,addr string,parser TcpParser) *TcpServer {
	if maxSocketNum == 0 || ioNum == 0 {
		logging.Error("NewTcpServer invalid config")
		return nil
	}

	return &TcpServer{
		ConnList:     make([]*ConnInfo, maxSocketNum),
		MaxSocketNum: maxSocketNum,
		IoThreadNum:  ioNum,
		Addr: addr,
		UniqueId: 0,
		Parser: parser,
		CheckTimeoutTs:checkTimeoutTs,
		TimeoutTs:timeoutTs,
	}
}

func (s *TcpServer) Start() error {

	for i := 0; i < s.IoThreadNum; i++ {
		iothread := NewIoThread(s,i)
		s.IoThreadList = append(s.IoThreadList, iothread)
		iothread.Start()
	}

	listenfd, err := s.CreateListenSocket(s.Addr)
	if err != nil {
		logging.Error("TcpServer CreateListenSocket failed")
		return err
	}

	for {
		//logging.Debug("loop1")
		fd, addr, err := syscall.Accept(listenfd)
		if err != nil {
			logging.Error("TcpServer Accept failed")
			continue
		}

		if fd > s.MaxSocketNum {
			logging.Error("IoThread Accept invalid fd:%d", fd)
			continue
		}
		id :=  s.CreateUniqueId()
		
		
		socketInfo := NewSocketInfo(fd,id, addr)
		ioIndex := fd % s.IoThreadNum
		logging.Debug("Accept fd=%d,id:%d,ioIndex=%d,addr=%+v,err=%+v", fd,id,ioIndex, addr, err)
		s.IoThreadList[ioIndex].Notify(EVENT_ACCEPT,socketInfo)
	}

	return nil
}

func (s *TcpServer)CreateUniqueId() uint64{
	s.UniqueId += 1
	return s.UniqueId
}

//异步场景下检查socket唯一id是否匹配
func (s *TcpServer) CheckSocketId(fd int,id uint64) bool {
	if fd > s.MaxSocketNum {
		logging.Error("TcpServer CheckSocketId failed,fd:%d,id:%d",fd,id)
		return false
	}
	c := s.ConnList[fd]
	if  c== nil {
		logging.Error("TcpServer CheckSocketId failed, already closed,fd:%d,id:%d",fd,id)
		return false
	}
	return c.SInfo.Id == id
}

func (s *TcpServer) SendMsg(fd int,id uint64, msg []byte) error {
	if len(msg) == 0 {
		logging.Error("TcpServer SendMsg empty,fd=%d", fd)
		return nil
	}
	if !s.CheckSocketId(fd,id){
		logging.Error("TcpServer SendMsg CheckSocketId failed,fd:%d,id:%d",fd,id)
		return errors.New("CheckSocketId failed")
	}
	c := s.ConnList[fd]
	if c != nil {
		c.AsynSendMsg(msg)
		logging.Debug("TcpServer SendMsg ok,msg:%#v,fd:%d,id:%d",msg,fd,id)
	}else{
		logging.Error("TcpServer SendMsg failed,socket closed,msg:%#v,fd:%d,id:%d",msg,fd,id)
	}
	return nil
}

func (s *TcpServer) CreateListenSocket(ipport string) (int, error) {
	socket, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	var flag = int(1)
	err := syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, flag)
	if err != nil {
		logging.Error("TcpServer Setsockopt failed")
		return 0, err
	}

	ipinfo := strings.Split(ipport, ":")
	if len(ipinfo) != 2 {
		logging.Error("TcpServer invalid ipport:%s", ipport)
		return 0, errors.New("invalid ipport")
	}

	ip, err := parseIPv4(ipinfo[0])
	if err != nil {
		logging.Error("TcpServer parseIPv4 failed")
		return 0, errors.New("invalid ip")
	}

	port, err := parsePort(ipinfo[1])
	if err != nil {
		logging.Error("TcpServer parsePort failed")
		return 0, err
	}

	addr := &syscall.SockaddrInet4{
		//Family: syscall.AF_INET,
		Port: port,
		Addr: ip,
	}

	err = syscall.Bind(socket, addr)
	if err != nil {
		logging.Error("TcpServer Bind failed")
		return 0, err
	}

	err = syscall.Listen(socket, ACCEPT_CHAN_LEN)
	if err != nil {
		logging.Error("TcpServer listen failed")
		return 0, err
	}
	return socket, nil
}
