package main
import (
    "fmt"
    "runtime"
    "sync"
    "time"
	"syscall"
)
const COUNT = 10000
func bench1(ch chan int) time.Duration {
    t := time.Now()
    for i := 0; i < COUNT; i++ {
        ch <- i
    }
    var v int
    for i := 0; i < COUNT; i++ {
        v = <-ch
    }
    _ = v
    return time.Now().Sub(t)
}
func bench2(s []int) time.Duration {
    t := time.Now()
    for i := 0; i < COUNT; i++ {
        s[i] = i
    }
    var v int
    for i := 0; i < COUNT; i++ {
        v = s[i]
    }
    _ = v
    return time.Now().Sub(t)
}
func bench3(s []int, mutex *sync.Mutex) time.Duration {
    t := time.Now()
    for i := 0; i < COUNT; i++ {
        mutex.Lock()
        s[i] = i
        mutex.Unlock()
    }
    var v int
    for i := 0; i < COUNT; i++ {
        mutex.Lock()
        v = s[i]
        mutex.Unlock()
    }
    _ = v
    return time.Now().Sub(t)
}

func bench4(efd int) time.Duration {
	wakezero := []byte{1, 0, 0, 0, 0, 0, 0, 0}
	t := time.Now()
	for i := 0; i < COUNT; i++ {
		//fmt.Printf("write i:%d\n",i)
        _,err := syscall.Write(efd, wakezero[:])
		if err != nil{
			fmt.Printf("Write failed,err:%s",err.Error())
		}
		
    }
    var data [8]byte
				
    for i := 0; i < COUNT; i++ {
		fmt.Printf("Read i:%d\n",i)
      n,err := syscall.Read(efd, data[:])
	  fmt.Printf(" Read,n:%d,%#v\n",n,data)
	  if err != nil{
			fmt.Printf("Read,err:%s",err.Error())
			return time.Now().Sub(t)
		}
	  
    }
	
	return time.Now().Sub(t)
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    ch := make(chan int, COUNT)
    s := make([]int, COUNT)
	
	r1, _, _ := syscall.Syscall(284, syscall.O_NONBLOCK, syscall.O_NONBLOCK, syscall.O_NONBLOCK)
	var efd = int(r1)
	syscall.SetNonblock(efd, true)
	
    var mutex sync.Mutex
    fmt.Println("channel\tslice\tmutex_slice\teventfd")
    for i := 0; i < 10; i++ {
        fmt.Printf("%v\t%v\t%v\t%v\n", bench1(ch), bench2(s), bench3(s, &mutex))
    }
}