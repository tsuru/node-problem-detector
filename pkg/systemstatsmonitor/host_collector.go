/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"github.com/shirou/gopsutil/v3/host"
	"k8s.io/klog/v2"

	ssmtypes "k8s.io/node-problem-detector/pkg/systemstatsmonitor/types"
	"k8s.io/node-problem-detector/pkg/util"
	"k8s.io/node-problem-detector/pkg/util/metrics"
)

type hostCollector struct {
	tags   map[string]string
	uptime *metrics.Int64Metric
}

func NewHostCollectorOrDie(hostConfig *ssmtypes.HostStatsConfig) *hostCollector {
	hc := hostCollector{map[string]string{}, nil}

	kernelVersion, err := host.KernelVersion()
	if err != nil {
		klog.Fatalf("Failed to retrieve kernel version: %v", err)
	}
	hc.tags["kernel_version"] = kernelVersion

	osVersion, err := util.GetOSVersion()
	if err != nil {
		klog.Fatalf("Failed to retrieve OS version: %v", err)
	}
	hc.tags["os_version"] = osVersion

	// Use metrics.Sum aggregation method to ensure the metric is a counter/cumulative metric.
	if hostConfig.MetricsConfigs["host/uptime"].DisplayName != "" {
		hc.uptime, err = metrics.NewInt64Metric(
			metrics.HostUptimeID,
			hostConfig.MetricsConfigs[string(metrics.HostUptimeID)].DisplayName,
			"The uptime of the operating system",
			"second",
			metrics.LastValue,
			[]string{"kernel_version", "os_version"})
		if err != nil {
			klog.Fatalf("Error initializing metric for host/uptime: %v", err)
		}
	}

	return &hc
}

func (hc *hostCollector) collect() {
	if hc == nil {
		return
	}

	uptime, err := host.Uptime()
	if err != nil {
		klog.Errorf("Failed to retrieve uptime of the host: %v", err)
		return
	}

	if hc.uptime != nil {
		_ = hc.uptime.Record(hc.tags, int64(uptime))
	}
}
