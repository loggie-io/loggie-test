/*
Copyright 2021 Loggie Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"time"

	"github.com/loggie-io/loggie/pkg/core/log"
)

func AsyncRunWithTimeout(f func(), timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	done := make(chan struct{})
	go func() {
		now := time.Now()
		f()
		since := time.Since(now)
		if since > timeout {
			log.Info("func run too long(%ds)", since/time.Second)
		}

		done <- struct{}{}
	}()
	for {
		select {
		case <-done:
			return
		case <-timer.C:
			return
		}
	}
}
