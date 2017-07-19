package libutil

import (
	"os"
)

var (
	dumpFlag   = os.O_CREATE | os.O_WRONLY
	dumpMode   = os.FileMode(0777)
	dumpPrefix = "panic."
)

func ReviewDumpPanic(file *os.File) error {
	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileinfo.Size() == 0 {
		file.Close()
		return os.Remove(file.Name())
	}
	return nil
}
