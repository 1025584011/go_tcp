// +build windows

package libutil

import (
	"os"
	"strconv"
	"syscall"
)

const (
	kernel32dll  = "kernel32.dll"
	stderrHandle = uint32(-12 & 0xFFFFFFFF)
)

func DumpPanic(suffix string) (*os.File, error) {
	filename := dumpPrefix + suffix + "." + strconv.Itoa(os.Getpid())
	file, err := os.OpenFile(filename, dumpFlag, dumpMode)
	if err != nil {
		return file, err
	}
	kernel32 := syscall.NewLazyDLL(kernel32dll)
	setStdHandle := kernel32.NewProc("SetStdHandle")
	v, _, err := setStdHandle.Call(uintptr(stderrHandle), file.Fd())
	if v == 0 {
		return file, err
	}
	return file, nil
}
