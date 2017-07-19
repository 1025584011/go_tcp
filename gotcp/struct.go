package gotcp

import (
	//"bytes"
	//"syscall"
)



const(
    EVENT_BEGIN = iota //0
    EVENT_ACCEPT //1
    EVENT_WRITE //2
	EVENT_CLOSE //3
)

type NotifyEvent struct {
	Type int //1.accept  2.write 3.close
	Info *SocketInfo
}