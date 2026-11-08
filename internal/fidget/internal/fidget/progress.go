package fidget

/*
#include <stdint.h>
*/
import "C"
import "sync"

var (
	progressMu   sync.Mutex
	progressFunc ProgressFunc
)

func setProgressFunc(fn ProgressFunc) {
	progressMu.Lock()
	progressFunc = fn
	progressMu.Unlock()
}

//export goProgressCallback
func goProgressCallback(completed C.uint, total C.uint) {
	progressMu.Lock()
	fn := progressFunc
	progressMu.Unlock()
	if fn != nil {
		fn(uint32(completed), uint32(total))
	}
}
