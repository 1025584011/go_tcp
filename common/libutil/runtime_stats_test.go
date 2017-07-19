package libutil

import (
	"strconv"
	"testing"
	"time"
)

func TestHumanSize(t *testing.T) {
	if humanSize(uint64(3)) != "3B" {
		t.Fail()
	}
	if humanSize(uint64(1331)) != "1.3KB" {
		t.Fail()
	}
	if humanSize(uint64(1363148)) != "1.3MB" {
		t.Fail()
	}
	if humanSize(uint64(1395864371)) != "1.3GB" {
		t.Fail()
	}
}

func TestRuntimeStats(t *testing.T) {
	go RuntimeStats(time.Second, nil)
	go func() {
		for {
			m := make(map[string]int)
			for i := 0; i < 1000; i++ {
				m[strconv.Itoa(i)] = i
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	time.Sleep(10 * time.Second)
}
