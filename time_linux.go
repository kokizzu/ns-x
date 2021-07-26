// +build cgo,!time_compiled,linux

package networksimulator

/*
#cgo CFLAGS: -I${SRCDIR}/cpp
#cgo LDFLAGS: -L${SRCDIR}/binary/linux -ltime -lstdc++
#include "cpp/library.h"
*/
import "C"
import "time"

func Now() time.Time {
	return time.Unix(0, int64(C.now()))
}
