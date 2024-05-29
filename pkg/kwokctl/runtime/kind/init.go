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

package kind

import (
	"sigs.k8s.io/kwok/pkg/consts"
	"sigs.k8s.io/kwok/pkg/kwokctl/runtime"
)

func init() {
	runtime.DefaultRegistry.Register(consts.RuntimeTypeKind, NewDockerCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeKindPodman, NewPodmanCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeKindNerdctl, NewNerdctlCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeKindLima, NewLimaCluster)
	runtime.DefaultRegistry.Register(consts.RuntimeTypeKindFinch, NewFinchCluster)
}
