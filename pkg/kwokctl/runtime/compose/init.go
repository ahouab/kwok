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

package compose

import (
	"sigs.k8s.io/kwok/pkg/consts"
	"sigs.k8s.io/kwok/pkg/kwokctl/runtime"
)

func init() {
	runtime.DefaultRegistry.Register(consts.RuntimeTypeDocker, NewDockerCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypePodman, NewPodmanCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeNerdctl, NewNerdctlCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeLima, NewLimaCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeFinch, NewFinchCluster)
}
