package main

import (
	"cmq_proj/proto"
	"common/logging"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"wx_project/common/cmqprotocol"

	"github.com/golang/protobuf/proto"
)

var ReadCount int64
var PreReadCount int64
var WriteCount int64
var PreWriteCount int64

var ClientList []*net.TCPConn
var clientCount int
var threadNum int

func RequestFunc(index int) {
	s := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s_b := []byte(s)
	b := make([]byte, 10000)
	for {
		i := 0
		for ; threadNum*i+index < clientCount; i++ {
			_, err1 := (*ClientList[threadNum*i+index]).Write(s_b)
			if err1 != nil {
				fmt.Printf("write failed,err1=%v,err2=%v", err1)
				continue
			}
		}
		i = 0
		for ; threadNum*i+index < clientCount; i++ {

			_, err2 := (*ClientList[threadNum*i+index]).Read(b)
			if err2 != nil {
				fmt.Printf("write Read,err1=%v,err2=%v", err2)
				continue
			}
			atomic.AddInt64(&WriteCount, 1)
		}

	}
}

func main() {
	args := os.Args
	if args == nil || len(args) < 3 {
		fmt.Println("usage: tcpserver +clientCount + serverAddr")
		return
	}

	go func() {
		err := http.ListenAndServe("0.0.0.0:54444", nil)
		if err != nil {
			logging.Error("ListenAndServe: %s", err.Error())
		}
	}()

	clientCount, _ = strconv.Atoi(args[1])
	/*
		tcpAddr, _ := net.ResolveTCPAddr("tcp4", args[2])
		ClientList = make([]*net.TCPConn, clientCount)

		var err error
		for i := 0; i < clientCount; i++ {
			ClientList[i], err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				fmt.Printf("CreateClient DialTCP failed")
				continue
			}
			ClientList[i].SetDeadline(1)
			ClientList[i].SetWriteDeadline(1)
		}
		threadNum = runtime.NumGoroutine()

		for i := 0; i < threadNum; i++ {
			go RequestFunc(i)
		}
	*/

	for i := 0; i < clientCount; i++ {
		go SynCreateClient(args[2])
	}

	go PrintQps()
	ChanShutdown := make(chan os.Signal)
	signal.Notify(ChanShutdown, syscall.SIGINT)
	<-ChanShutdown
}

func PrintQps() {
	for {
		time.Sleep(time.Second * 10)
		qpsRead := (ReadCount - PreReadCount) / 10
		fmt.Printf("qpsRead =%f\n", qpsRead)
		PreReadCount = ReadCount

		qpsWrite := (WriteCount - PreWriteCount) / 10
		fmt.Printf("qpsWrite =%f\n", qpsWrite)
		PreWriteCount = WriteCount
	}

}

func ReadThread(conn net.Conn) {
	for {
		b := make([]byte, 10000)
		conn.Read(b)
		//count, err := conn.Read(b)
		//fmt.Printf("Read count=%d,err=%v\n", count, err)
		atomic.AddInt64(&ReadCount, 1)
	}
}

func AsynCreateClient(serverAddr string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("CreateClient DialTCP failed")
		return
	}
	go ReadThread(conn)
	s := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for {
		conn.Write([]byte(s))
		//count, err := conn.Write([]byte(s))
		//fmt.Printf("Write count=%d,err=%v\n", count, err)
		atomic.AddInt64(&WriteCount, 1)
	}

}

func SynCreateClient(serverAddr string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("CreateClient DialTCP failed")
		return
	}
	//go ReadThread(conn)
	s := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s_b := []byte(s)
	//s := GetSendData()
	b := make([]byte, 10000)
	for {
		_, err1 := conn.Write(s_b)
		_, err2 := conn.Read(b)
		if err1 != nil || err2 != nil {
			fmt.Printf("read or write failed,err1=%v,err2=%v", err1, err2)
			return
		}
		//count, err := conn.Write([]byte(s))
		//fmt.Printf("Write count=%d,err=%v\n", count, err)
		atomic.AddInt64(&WriteCount, 1)
		//WriteCount++
	}

}

func GetSendData() []byte {

	base := &User.BaseMessage_CS{}
	base.Type = User.Null_TickRequestNull_CS.Enum()
	send := &User.TickRequestNull_CS{}
	//send.Requesttime = rev.Requesttime
	//send.Mytime = proto.Uint32(uint32(time.Now().Unix()))

	data, _ := proto.Marshal(send)

	base.Data = data
	data, _ = proto.Marshal(base)

	data2 := cmqprotocol.NewCmqPacket(data).Serialize()
	return data2
}
