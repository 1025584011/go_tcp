package http

import(
	"go_tcp/gotcp"
	"go_tcp/common/logging"
	"strings"
)


type HttpParser struct{

}

func(p *HttpParser)Unpack(msg []byte,c *gotcp.ConnInfo)(ok bool,packlen int){
	//logging.Debug("HttpParser Unpack,msg:%#v",msg)
	s := string(msg)
	index := strings.Index(s,"\r\n\r\n")
	if index == -1 {
		logging.Debug("HttpParser not find end")
		return true,0
	}else{
		logging.Debug("HttpParser find end")
		return true,index+4
	}
}

func(p *HttpParser)HandlePack(msg []byte,c *gotcp.ConnInfo)(ok bool){
	//logging.Debug("HttpParser HandlePack,msg:%s",string(msg))
	r := "HTTP/1.1 200 OK\r\nDate: Tue, 18 Jul 2017 09:49:30 GMT\r\nContent-Length: 4\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nabcd"
	c.SynSendMsg([]byte(r))
	c.SynClose()
	//c.Close()
	return true
}

