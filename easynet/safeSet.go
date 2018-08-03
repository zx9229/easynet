package easynet

import (
	"sync"
)

type safeSet struct {
	cache map[*EasySocket]bool
	mutex sync.Mutex
}

func newSafeSet() *safeSet {
	curData := new(safeSet)
	curData.cache = make(map[*EasySocket]bool)
	return curData
}

func (thls *safeSet) Add(sock *EasySocket) {
	thls.cache[sock] = true
}

func (thls *safeSet) Del(sock *EasySocket) {
	delete(thls.cache, sock)
}
