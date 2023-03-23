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

package internalversion

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KwokConfiguration provides configuration for the Kwok.
type KwokConfiguration struct {
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta
	// Options holds information about the default value.
	Options KwokConfigurationOptions
}

// KwokConfigurationOptions holds information about the options.
type KwokConfigurationOptions struct {
	// The default IP assigned to the Pod on maintained Nodes.
	CIDR string

	// The ip of all nodes maintained by the Kwok
	NodeIP string

	// The name of all nodes maintained by the Kwok
	NodeName string

	// The port of all nodes maintained by the Kwok
	NodePort int

	// TLSCertFile is the file containing x509 Certificate
	TLSCertFile string

	// TLSPrivateKeyFile is the ile containing x509 private key
	TLSPrivateKeyFile string

	// Default option to manage (i.e., maintain heartbeat/liveness of) all Nodes or not.
	ManageAllNodes bool

	// Default annotations specified on Nodes to demand manage.
	ManageNodesWithAnnotationSelector string

	// Default labels specified on Nodes to demand manage.
	ManageNodesWithLabelSelector string

	// If a Node/Pod is on a managed Node and has this annotation status will not be modified
	DisregardStatusWithAnnotationSelector string

	// If a Node/Pod is on a managed Node and has this label status will not be modified
	DisregardStatusWithLabelSelector string

	// ServerAddress is server address of the Kwok.
	ServerAddress string

	// Experimental support for getting pod ip from CNI, for CNI-related components, Only works with Linux.
	EnableCNI bool

	// enableDebuggingHandlers enables server endpoints for log collection
	// and local running of containers and commands
	EnableDebuggingHandlers bool

	// enableContentionProfiling enables lock contention profiling, if enableDebuggingHandlers is true.
	EnableContentionProfiling bool

	// EnableProfiling enables /debug/pprof handler.
	EnableProfilingHandler bool

	// TotalParallel is the total number of parallel tasks.
	TotalParallel int

	// NodeParallelPriority is the priority of parallel tasks for Node.
	NodeParallelPriority int

	// NodeDelayParallelPriority is the priority of parallel delay tasks for Node.
	NodeDelayParallelPriority int

	// PodParallelPriority is the priority of parallel tasks for Pod.
	PodParallelPriority int

	// PodDelayParallelPriority is the priority of parallel delay tasks for Pod.
	PodDelayParallelPriority int
}
