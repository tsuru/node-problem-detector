//go:build unix

/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package systemstatsmonitor

import (
	"github.com/prometheus/procfs"
	"k8s.io/klog/v2"
)

func (mc *memoryCollector) collect() {
	if mc == nil {
		return
	}

	proc, err := procfs.NewDefaultFS()
	if err != nil {
		klog.Errorf("Failed to find /proc mount point: %v", err)
		return
	}
	meminfo, err := proc.Meminfo()
	if err != nil {
		klog.Errorf("Failed to retrieve memory stats: %v", err)
		return
	}

	if mc.mBytesUsed != nil {
		if meminfo.MemFree != nil {
			_ = mc.mBytesUsed.Record(map[string]string{stateLabel: "free"}, int64(*meminfo.MemFree)*1024)
		}
		if meminfo.Buffers != nil {
			_ = mc.mBytesUsed.Record(map[string]string{stateLabel: "buffered"}, int64(*meminfo.Buffers)*1024)
		}
		if meminfo.Cached != nil {
			_ = mc.mBytesUsed.Record(map[string]string{stateLabel: "cached"}, int64(*meminfo.Cached)*1024)
		}
		if meminfo.Slab != nil {
			_ = mc.mBytesUsed.Record(map[string]string{stateLabel: "slab"}, int64(*meminfo.Slab)*1024)
		}
		if meminfo.MemTotal != nil && meminfo.MemFree != nil && meminfo.Buffers != nil && meminfo.Cached != nil && meminfo.Slab != nil {
			memUsed := *meminfo.MemTotal - *meminfo.MemFree - *meminfo.Buffers - *meminfo.Cached - *meminfo.Slab
			_ = mc.mBytesUsed.Record(map[string]string{stateLabel: "used"}, int64(memUsed)*1024)
		}
	}

	if mc.mPercentUsed != nil && meminfo.MemTotal != nil && *meminfo.MemTotal > 0 &&
		meminfo.MemFree != nil && meminfo.Buffers != nil && meminfo.Cached != nil && meminfo.Slab != nil {
		ratio := float64(*meminfo.MemTotal-*meminfo.MemFree-*meminfo.Buffers-*meminfo.Cached-*meminfo.Slab) / float64(*meminfo.MemTotal)
		_ = mc.mPercentUsed.Record(map[string]string{stateLabel: "used"}, float64(ratio*100.0))
	}

	if mc.mDirtyUsed != nil {
		if meminfo.Dirty != nil {
			_ = mc.mDirtyUsed.Record(map[string]string{stateLabel: "dirty"}, int64(*meminfo.Dirty)*1024)
		}
		if meminfo.Writeback != nil {
			_ = mc.mDirtyUsed.Record(map[string]string{stateLabel: "writeback"}, int64(*meminfo.Writeback)*1024)
		}
	}

	if mc.mAnonymousUsed != nil {
		if meminfo.ActiveAnon != nil {
			_ = mc.mAnonymousUsed.Record(map[string]string{stateLabel: "active"}, int64(*meminfo.ActiveAnon)*1024)
		}
		if meminfo.InactiveAnon != nil {
			_ = mc.mAnonymousUsed.Record(map[string]string{stateLabel: "inactive"}, int64(*meminfo.InactiveAnon)*1024)
		}
	}

	if mc.mPageCacheUsed != nil {
		if meminfo.ActiveFile != nil {
			_ = mc.mPageCacheUsed.Record(map[string]string{stateLabel: "active"}, int64(*meminfo.ActiveFile)*1024)
		}
		if meminfo.InactiveFile != nil {
			_ = mc.mPageCacheUsed.Record(map[string]string{stateLabel: "inactive"}, int64(*meminfo.InactiveFile)*1024)
		}
	}

	if mc.mUnevictableUsed != nil && meminfo.Unevictable != nil {
		_ = mc.mUnevictableUsed.Record(map[string]string{}, int64(*meminfo.Unevictable)*1024)
	}
}
