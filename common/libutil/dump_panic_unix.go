// +build !windows

package libutil

import (
	"os"
	"strconv"
	"syscall"
)

func DumpPanic(suffix string) (*os.File, error) {
	filename := dumpPrefix + suffix + "." + strconv.Itoa(os.Getpid())
	file, err := os.OpenFile(filename, dumpFlag, dumpMode)
	if err != nil {
		return file, err
	}
	if err := syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		return file, err
	}
	return file, nil
}
