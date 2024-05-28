/*
Copyright 2022 The Kubernetes Authors.

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

package k8s

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var lockEnabled = map[string]bool{}

// GetFeatureGates returns the feature gates for the given version
func GetFeatureGates(version int) string {
	if version < 0 {
		return ""
	}
	// Enable only the beta feature of the final GA
	isGA := map[string]bool{}
	for _, raw := range rawData {
		if raw.Stage == GA {
			_, ok := isGA[raw.Name]
			if !ok {
				isGA[raw.Name] = true
			}
		} else if raw.Stage == Deprecated {
			isGA[raw.Name] = false
		}
	}

	enables := map[string]bool{}
	for _, raw := range rawData {
		if raw.Contain(version) {
			if raw.Stage == Beta {
				enables[raw.Name] = isGA[raw.Name] || lockEnabled[raw.Name]
			}
		}
	}

	gates := make([]string, 0, len(enables))
	for name, enable := range enables {
		gates = append(gates, name+"="+strconv.FormatBool(enable))
	}
	sort.Strings(gates)
	return strings.Join(gates, ",")
}

// FeatureSpec is the specification of a feature
type FeatureSpec struct {
	Name  string
	Stage Stage
	Since int
	Until int
}

// Contain returns true if the version is in the range of the feature
func (f *FeatureSpec) Contain(v int) bool {
	return f.Since <= v &&
		(f.Until < 0 || v <= f.Until)
}

// Verification of the data
func (f *FeatureSpec) Verification() error {
	if f.Since < 0 {
		return fmt.Errorf("invalid since: %d", f.Since)
	}
	if f.Until >= 0 && f.Until < f.Since {
		return fmt.Errorf("invalid until: %d < since: %d", f.Until, f.Since)
	}
	return nil
}

// Stage is the stage of a feature.
type Stage string

// The following stages are defined in https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
const (
	Alpha = Stage("ALPHA")
	Beta  = Stage("BETA")
	GA    = Stage("GA")

	// Deprecated
	Deprecated = Stage("DEPRECATED")
)
