package main

import (
	"cmq_proj/proto"
	"common/goredis"
	"fmt"
	"net"
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

var Redis_handle *goredis.Redis

func main() {
	args := os.Args
	if args == nil || len(args) < 3 {
		fmt.Println("usage: tcpserver +clientCount + serverAddr")
		return
	}

	clientCount, _ := strconv.Atoi(args[1])

	for i := 0; i < clientCount; i++ {
		go CreateRedisClient(int(i), args[2])
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

func CreateRedisClient(i int, serverAddr string) {
	Redis_handle, e := goredis.DialURL("tcp://" + serverAddr + "/0?timeout=10s&maxidle=10")
	if e != nil {
		fmt.Printf("Redis_handle DialURL failed,err=%v", e)
		return
	}
	key := strconv.Itoa(i) + "aaa"
	for {
		_, err := Redis_handle.Get(key)
		if err != nil {
			fmt.Printf("Redis_handle Get failed,err=%v", err)
			return
		}
		atomic.AddInt64(&WriteCount, 1)
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
	//s := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	s := GetSendData()
	b := make([]byte, 10000)
	for {
		_, err1 := conn.Write([]byte(s))
		_, err2 := conn.Read(b)
		if err1 != nil || err2 != nil {
			fmt.Printf("read or write failed,err1=%v,err2=%v", err1, err2)
			return
		}
		//count, err := conn.Write([]byte(s))
		//fmt.Printf("Write count=%d,err=%v\n", count, err)
		atomic.AddInt64(&WriteCount, 1)
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
