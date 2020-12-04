package avfoundation

// extern void onData(void*, void*, int);
import "C"
import (
	"sync"
	"unsafe"
)

var mu sync.Mutex
var nextID handleID

type dataCb func(data []byte)

var handles = make(map[handleID]dataCb)

type handleID int

//export onData
func onData(userData unsafe.Pointer, buf unsafe.Pointer, length C.int) {
	data := C.GoBytes(buf, length)

	handleNum := (*C.int)(userData)
	cb, ok := lookup(handleID(*handleNum))
	if ok {
		cb(data)
	}
}

func register(fn dataCb) handleID {
	mu.Lock()
	defer mu.Unlock()

	nextID++
	for handles[nextID] != nil {
		nextID++
	}
	handles[nextID] = fn

	return nextID
}

func lookup(i handleID) (cb dataCb, ok bool) {
	mu.Lock()
	defer mu.Unlock()

	cb, ok = handles[i]
	return
}

func unregister(i handleID) {
	mu.Lock()
	defer mu.Unlock()

	delete(handles, i)
}
