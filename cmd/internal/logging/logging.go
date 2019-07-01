// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"github.com/xgfone/go-tools/lifecycle"
	"github.com/xgfone/klog"
)

// Init initializes the logging.
func Init(level, filepath string) error {
	if filepath != "" {
		file, err := klog.NewSizedRotatingFile(filepath, 100*1024*1024, 100)
		if err != nil {
			return err
		}
		lifecycle.Register(func() { file.Close() })
		klog.Std = klog.Std.WithWriter(klog.StreamWriter(file))
	}

	klog.Std = klog.Std.WithLevel(klog.NameToLevel(level)).WithKv("caller", klog.Caller())
	klog.AppendCleaner(lifecycle.Stop)
	return nil
}
