//go:build tools
// +build tools

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

// Package tools is used to track binary dependencies with go modules
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	// code-generator
	_ "k8s.io/code-generator"

	// controller-gen
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"

	// gen-crd-api-reference-docs
	_ "github.com/ahmetb/gen-crd-api-reference-docs"

	// shfmt
	_ "mvdan.cc/sh/v3/cmd/shfmt"

	// misspell
	_ "github.com/client9/misspell/cmd/misspell"

	// kustomize
	_ "sigs.k8s.io/kustomize/kustomize/v5"
)
