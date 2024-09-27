// Copyright 2024 github.com/lvan100
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"sync/atomic"
	"time"
)

// now is the cached timestamp with microsecond precision.
var now int64

func init() {
	start := make(chan struct{})
	go func() {
		atomic.StoreInt64(&now, UnixMicro(time.Now()))
		start <- struct{}{}
		// update the timestamp every one millisecond.
		for t := range time.Tick(time.Millisecond) {
			atomic.StoreInt64(&now, UnixMicro(t))
		}
	}()
	<-start
}

// MicroNow returns the cached current Unix timestamp in microseconds.
func MicroNow() int64 {
	return atomic.LoadInt64(&now)
}

// UnixMicro converts the given Time to Unix timestamp in microseconds.
func UnixMicro(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}
