// It includes skill, equipment, card and so on
package main

import (
	"flag"
	"fmt"
	"time"

	"net/http"
	"runtime"

	"go_tcp/common/libutil"
	"go_tcp/common/logging"
	_ "net/http/pprof"
	"go_tcp/gotcp"
	http_p "go_tcp/protocol/http"
)

var cfg = struct {
	Log struct {
		File   string
		Level  string
		Name   string
		Suffix string
	}

	Prog struct {
		CPU        int
		Daemon     bool
		HealthPort string
	}

	Server struct {
		Redis    string
		Mysql    string
		ListenAddr string
		MaxSocket int
		IoNum int    //配置为 core个数-1时性能最佳,不能超过core个数，会出问题
	}
}{}

/*
demo功能点:
1.日志
2.mysql
3.redis
4.http框架
5.json
6.http客户端
7.配置解析
*/

//注册http回调
func registerHttpHandle() {
	//http.HandleFunc("/test", HandlerTest)
}

func main() {
	//配置解析
	config := flag.String("c", "conf/config.json", "config file")
	flag.Parse()
	if err := libutil.ParseJSON(*config, &cfg); err != nil {
		fmt.Printf("parse config %s error: %s\n", *config, err.Error())
		return
	}

	//日志
	/*
		if err := libutil.TRLogger(cfg.Log.File, cfg.Log.Level, cfg.Log.Name, cfg.Log.Suffix, cfg.Prog.Daemon); err != nil {
			fmt.Printf("init time rotate logger error: %s\n", err.Error())
			return
		}
	*/
	if cfg.Prog.CPU == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU()) //配0就用所有核
	} else {
		runtime.GOMAXPROCS(cfg.Prog.CPU)
	}

	//日志
	logging.Debug("server start")

	//Mysql
	//InitMysql()

	//Redis
	//InitRedis()

	go func() {
		err := http.ListenAndServe(cfg.Prog.HealthPort, nil)
		if err != nil {
			logging.Error("ListenAndServe: %s", err.Error())
		}
	}()

	gotcp.InitServer(cfg.Server.IoNum, cfg.Server.MaxSocket,cfg.Server.ListenAddr,&http_p.HttpParser{})

	libutil.InitSignal()

	//registerHttpHandle()
	//httpclient
	//TestHttpClient()

	go func() {
		//http.ListenAndServe(cfg.Server.PortInfo, nil)
	}()

	file, err := libutil.DumpPanic("gsrv")
	if err != nil {
		logging.Error("init dump panic error: %s", err.Error())
	}

	defer func() {
		logging.Info("server stop...:%d", runtime.NumGoroutine())
		time.Sleep(time.Second)
		logging.Info("server stop...,ok")
		if err := libutil.ReviewDumpPanic(file); err != nil {
			logging.Error("review dump panic error: %s", err.Error())
		}

	}()
	<-libutil.ChanRunning

}
