/*
Copyright 2023 The Kubernetes Authors.

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

package binary

import (
	"bytes"
	"fmt"
	"text/template"

	_ "embed"
)

//go:embed metrics_server_apiservice.yaml.tpl
var metricsServerAPIServiceYamlTpl string

var metricsServerAPIServiceYamlTemplate = template.Must(template.New("_").Parse(metricsServerAPIServiceYamlTpl))

// BuildMetricsServerAPIService builds the metrics server apiservice yaml content.
func BuildMetricsServerAPIService(conf BuildMetricsServerAPIServiceConfig) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := metricsServerAPIServiceYamlTemplate.Execute(buf, conf)
	if err != nil {
		return "", fmt.Errorf("failed to execute metrics server apiservice yaml template: %w", err)
	}
	return buf.String(), nil
}

// BuildMetricsServerAPIServiceConfig is the config for BuildMetricsServerAPIService.
type BuildMetricsServerAPIServiceConfig struct {
	Port uint32
}
