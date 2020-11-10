// Copyright 2017 The Prometheus Authors
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

// A prometheus collector to execute custom shell script which have a formated output.

package collector

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	customScript = "custom_script"
)

var (
	// Set cmd line flag and defalut custom script path
	scriptPath = kingpin.Flag("collector.customscript.scriptPath", "custom scripts path").Default("/opt/prometheus/customScript").String()
)

type customScriptCollector struct {
	scriptPath string
	logger     log.Logger
}

func init() {
	// registerCollector("custom_script", defaultEnabled, NewCustomScriptCollector)
	registerCollector("custom_script", defaultDisabled, NewCustomScriptCollector)
}

// NewCustomScriptCollector returns a new Collector exposing custom script output stats.
func NewCustomScriptCollector(logger log.Logger) (Collector, error) {
	return &customScriptCollector{
		scriptPath: *scriptPath,
		logger:     logger,
	}, nil
}

func (c *customScriptCollector) runScript(script string) (result string, err error) {
	cmd := exec.Command("bash", "-c", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}
	result, err = out.String(), nil
	return
}

func (c *customScriptCollector) getCustomScriptOutput() (map[string]float64, error) {
	var customScriptOutput = map[string]float64{}

	scripts, err := filepath.Glob(c.scriptPath + "/*")
	if err != nil {
		return nil, err
	}
	level.Debug(c.logger).Log("msg", "get custom script", "path", c.scriptPath)

	// format script output: echo "key=value"
	for i := range scripts {
		script := scripts[i]
		if err := os.Chmod(script, 0755); err != nil {
			level.Error(c.logger).Log("msg", "chmod +x failed", "path", c.scriptPath, "script", script)
		}
		result, err := c.runScript(script)
		if err != nil || result == "" {
			level.Error(c.logger).Log("msg", "exec script failed", "script", script, "result", result, "error", err)
			continue
		}
		level.Debug(c.logger).Log("msg", "exec custom script", "script", script)

		// format unformated metric name
		k, v := strings.Replace(strings.Split(result, "=")[0], " ", "_", -1), strings.Split(result, "=")[1]
		level.Debug(c.logger).Log("msg", "script field", "key", k, "value", v)
		fv, err := strconv.ParseFloat(strings.Replace(v, "\n", "", -1), 64)
		if err != nil {
			level.Error(c.logger).Log("msg", "strconv failed", "value", v, "error", err)
			return nil, err
		}

		customScriptOutput[k] = fv
	}
	return customScriptOutput, nil
}

func (c *customScriptCollector) Update(ch chan<- prometheus.Metric) error {
	var metricType prometheus.ValueType

	// skip panic when exec unformated script output
	defer func() {
		if err := recover(); err != nil {
			level.Error(c.logger).Log("msg", "unsupported and unformatted custom script output", "error", err)
		}
	}()

	customScriptOutput, err := c.getCustomScriptOutput()
	if err != nil {
		return fmt.Errorf("couldn't get custom script output: %w", err)
	}

	metricType = prometheus.GaugeValue
	for k, v := range customScriptOutput {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, customScript, k),
				fmt.Sprintf("CustomScript information field %s.", k),
				nil, nil,
			),
			metricType, v,
		)
	}
	return nil
}
