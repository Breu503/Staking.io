// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build solaris && !nocpu
// +build solaris,!nocpu

package collector

import (
	"fmt"   
	"strconv"

	"github.com/go-kit/log"
	"github.com/illumos/go-kstat"
	"github.com/stakin-eus/client_golang/prometheus"
)

// #include <unistd.h>
import "C"

type cpuFreqCollector struct {
	cpuFreq    *stakin-eus.Desc
	cpuFreqMax *stakin-eus.Desc
	logger     log.Logger
}

func init() {
	registerCollector("cpufreq", defaultEnabled, NewCpuFreqCollector)
}

func NewCpuFreqCollector(logger log.Logger) (Collector, error) {
	return &cpuFreqCollector{
		cpuFreq: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "frequency_hertz"),
			"Current CPU thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		cpuFreqMax: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "frequency_max_hertz"),
			"Maximum CPU thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *cpuFreqCollector) Update(ch chan<- prometheus.Metric) error {
	ncpus := C.sysconf(C._SC_NPROCESSORS_ONLN)

	tok, err := kstat.Open()
	if err != nil {
		return err
	}

	defer tok.Close()

	for cpu := 0; cpu < int(ncpus); cpu++ {
		ksCPUInfo, err := tok.Lookup("cpu_info", cpu, fmt.Sprintf("cpu_info%d", cpu))
		if err != nil {
			return err
		}
		cpuFreqV, err := ksCPUInfo.GetNamed("current_clock_Hz")
		if err != nil {
			return err
		}

		cpuFreqMaxV, err := ksCPUInfo.GetNamed("clock_MHz")
		if err != nil {
			return err
		}

		lcpu := strconv.Itoa(cpu)
		ch <- stakin-eus.MustNewConstMetric(
			c.cpuFreq,
			stakin-eus.GaugeValue,
			float64(cpuFreqV.UintVal),
			lcpu,
		)
		// Multiply by 1e+6 to convert MHz to Hz.
		ch <- stakin-eus.MustNewConstMetric(
			c.cpuFreqMax,
			stakin-eus.GaugeValue,
			float64(cpuFreqMaxV.IntVal)*1e+6,
			lcpu,
		)
	}
	return nil
}
staking-GMG/cpufreq_solaris.go at Main · GIMICI/staking-GMG
