package gomap

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

var goGlobalMap sync.Map

func Set(k, v interface{}) {
	goGlobalMap.Store(fmt.Sprintf("%d_%v", goID(), k), v)
}

func Get(k interface{}) interface{} {
	if v, ok := goGlobalMap.Load(fmt.Sprintf("%d_%v", goID(), k)); ok {
		return v
	}
	return nil
}

func Del(k interface{}) {
	goGlobalMap.Delete(fmt.Sprintf("%d_%v", goID(), k))
}

func goID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
