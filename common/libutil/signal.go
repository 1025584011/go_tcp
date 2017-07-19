package libutil

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	ChanShutdown = make(chan os.Signal) //关闭信号chan
	ChanReload   = make(chan os.Signal) //-HUP信号chan
	ChanRunning  = make(chan bool)      //
	ChanHup      = make(chan os.Signal) //关闭信号chan
)

//初始化系统信号处理
//接收信号主要用来关闭时存档和动态数据加载
func InitSignal() {
	signal.Notify(ChanShutdown, syscall.SIGINT)
	signal.Notify(ChanShutdown, syscall.SIGTERM)
	signal.Notify(ChanHup, syscall.SIGHUP)
	//signal.Notify(ChanReload, syscall.SIGUSR1)
	InitSignalHandle()
}

func InitSignalHandle() {
	go func() {

		for {
			select {

			case <-ChanShutdown:
				ChanRunning <- false
			case <-ChanHup: //不处理终端关闭信号

			}
		}
	}()
}
