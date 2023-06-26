package sub

import (
	"sync/atomic"
	"time"
)

// Clock stores the time value as milliseconds since epoch
// with an accuracy of a microsecond.
//
// This time value does contain any time zone information.
func (s *Space) clock() {
	for t := range time.Tick(time.Microsecond) {
		atomic.StoreInt64(&s.now, t.UnixMilli())
	}
}
