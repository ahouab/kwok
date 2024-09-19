/*
Copyright 2024 The Kubernetes Authors.

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

// +k8s:deepcopy-gen=package
// +k8s:defaulter-gen=TypeMeta
// +groupName=operator.kwok.x-k8s.io

// +kubebuilder:rbac:groups="",resources=pods,verbs=create;delete;get
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments/scale,verbs=update

// Package v1alpha1 implements the v1alpha1 apiVersion of kwok's operator
package v1alpha1
